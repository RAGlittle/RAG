package qa

import (
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	InputKey  = "input"
	OutputKey = "text"
)

func NewQA(
	llm llms.Model,
	retriever schema.Retriever,
	memory schema.Memory,
) chains.ConversationalRetrievalQA {
	combineDocumentsChain := chains.LoadStuffQA(llm)
	// !! The condenseQuestionChain is a base LLM chain that always uses the output key "text"
	// !! therefore we need to set the output key to "text" in the QA chain
	condenseQuestionChain := chains.LoadCondenseQuestionGenerator(llm)

	return chains.ConversationalRetrievalQA{
		Memory:                memory,
		Retriever:             retriever,
		RephraseQuestion:      true,
		CombineDocumentsChain: combineDocumentsChain,
		CondenseQuestionChain: condenseQuestionChain,
		//FIXME: there is a bug with custom input / output keys in this chain
		InputKey: InputKey,
		// The inner LLM chains above
		OutputKey:               OutputKey,
		ReturnGeneratedQuestion: false,
		ReturnSourceDocuments:   false,
	}
}
