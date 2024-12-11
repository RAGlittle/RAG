package prompts

import (
	"github.com/tmc/langchaingo/prompts"
)

func NewQATemplate() prompts.ChatPromptTemplate {
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			"Answer the question based only on the following context : {{.context}}",
			[]string{
				"context",
			},
		),
		prompts.NewHumanMessagePromptTemplate(
			"Answer the question based on the above context: {{.question}}",
			[]string{
				"question",
			},
		),
	})
	return prompt
}
