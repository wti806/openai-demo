package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/wti806/openai-demo/go-server/internal/store"
)

type Request struct {
	SessionID string `json:"sessionId,omitempty"`
	Query     string `json:"query,omitempty"`
}

func decodeRequest(r *http.Request) (*Request, error) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

type DetectIntentHandler struct {
	OpenAIClient  *openai.Client
	SessionStore  *store.SessionStore
	FunctionStore *store.FunctionStore
	AssistantID   string
}

func (h *DetectIntentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest(r)
	if err != nil {
		log.Printf("fail to decode request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// create or get an existing thread
	threadID, err := h.getOrCreateThreadID(r.Context(), req.SessionID)
	if err != nil {
		log.Printf("fail to get or create thread: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	h.processMessage(r.Context(), threadID, req.Query)
}

func (h *DetectIntentHandler) getOrCreateThreadID(ctx context.Context, sessionID string) (string, error) {
	threadID, exists := h.SessionStore.GetSession(sessionID)
	if exists {
		return threadID, nil
	}
	threadResp, err := h.OpenAIClient.CreateThread(ctx, openai.ThreadRequest{
		Messages: []openai.ThreadMessage{},
		Metadata: map[string]any{},
	})
	if err != nil {
		return "", err
	}
	log.Printf("thread resp: %v", threadResp)
	h.SessionStore.AddSession(sessionID, threadResp.ID)
	return threadResp.ID, nil

}

func (h *DetectIntentHandler) processMessage(ctx context.Context, threadID string, query string) (string, error) {
	// 1. add message
	// TODO(wti)
	// 2. run thread
	runResp, err := h.OpenAIClient.CreateRun(ctx, threadID, openai.RunRequest{
		AssistantID: h.AssistantID,
	})
	if err != nil {
		return "", err
	}
	log.Printf("created run: %v", runResp)

	if err != nil {
		return "", err
	}
	return h.pollAndProcess(ctx, threadID, runResp.ID)
}

func (h *DetectIntentHandler) pollAndProcess(ctx context.Context, threadID, runID string) (string, error) {
	run, err := h.pollRun(ctx, threadID, runID)
	if err != nil {
		return "", err
	}
	log.Printf("run: %v", *run)

	return h.processRunResult(ctx, threadID, run)
}

func (h *DetectIntentHandler) pollRun(ctx context.Context, threadID, runID string) (*openai.Run, error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping the poller.")
			return nil, ctx.Err()
		case <-ticker.C:
			run, err := h.OpenAIClient.RetrieveRun(ctx, threadID, runID)
			if err != nil {
				return nil, err
			}
			switch run.Status {
			case openai.RunStatusCompleted:
			case openai.RunStatusRequiresAction:
				return &run, nil
			case openai.RunStatusFailed:
			case openai.RunStatusExpired:
			case openai.RunStatusCancelling:
				return nil, fmt.Errorf("unexpected run status: %v", run)
			}
		}
	}
}

func (h *DetectIntentHandler) processRunResult(ctx context.Context, threadID string, run *openai.Run) (string, error) {
	switch run.Status {
	case openai.RunStatusCompleted:
		// 3. list message
		// TODO(wti)
		return "", nil
	case openai.RunStatusRequiresAction:
		// 4. submit function call
		toolCalls := run.RequiredAction.SubmitToolOutputs.ToolCalls
		for _, call := range toolCalls {
			funcName := call.Function.Name
			funcParam := call.Function.Arguments
			outputStr, err := h.FunctionStore.ExecFunc(funcName, funcParam)
			if err != nil {
				return "", err
			}
			_, err = h.OpenAIClient.SubmitToolOutputs(ctx, threadID, run.ID, openai.SubmitToolOutputsRequest{
				ToolOutputs: []openai.ToolOutput{
					{
						ToolCallID: call.ID,
						Output:     outputStr,
					},
				},
			})
			if err != nil {
				return "", err
			}
		}
		return h.pollAndProcess(ctx, threadID, run.ID)
	default:
		return "", fmt.Errorf("invalid status")
	}
}
