package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	cx "cloud.google.com/go/dialogflow/cx/apiv3"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	openai "github.com/sashabaranov/go-openai"
	"github.com/wti806/openai-demo/go-server/internal/function"
	"github.com/wti806/openai-demo/go-server/internal/handler"
	"github.com/wti806/openai-demo/go-server/internal/store"
	"google.golang.org/api/option"
)

func main() {
	dfClient, err := initDialogflowClient()
	if err != nil {
		log.Fatalf("failed to init dialogflow client: %v/n", err)
	}
	defer dfClient.Close()

	openAIClient := initOpenAIClient()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Language", "Origin", "Access-Control-Request-Method", "X-Requested-With", "Content-Type"}),
	)

	r := mux.NewRouter()
	r.Handle("/detectIntent", &handler.DetectIntentHandler{
		OpenAIClient:  openAIClient,
		SessionStore:  store.NewSessionStore(),
		FunctionStore: initFunctionStore(),
		AssistantID:   os.Getenv("OPENAI_ASSISTANT_ID"),
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), corsHandler(r)))
}

func initDialogflowClient() (*cx.SessionsClient, error) {
	var options []option.ClientOption
	if env := os.Getenv("ENV"); env == "dev" {
		options = append(options, option.WithEndpoint("test-dialogflow.sandbox.googleapis.com:443"))
	}

	ctx := context.Background()
	return cx.NewSessionsClient(ctx, options...)
}

func initOpenAIClient() *openai.Client {
	return openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

func initFunctionStore() *store.FunctionStore {
	funcStore := store.NewFunctionStore()
	funcStore.RegisterFunc("resolve_order", function.ResolveOrder,
		reflect.TypeOf(function.Order{}),
		reflect.TypeOf(function.ResolveResp{}),
	)
	return funcStore
}
