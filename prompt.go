package internal

import (
	"bytes"
	"text/template"
)

// renderTemplate renders a template string with the provided data.
func renderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("example").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// one chunk translation
const (
	oneChunkInitialTranslationSystemMessage = `You are an expert linguist, specializing in translation from {{.sourceLang}} to {{.targetLang}}.`
	oneChunkInitialTranslationPrompt        = `This is an {{.sourceLang}} to {{.targetLang}} translation, please provide the {{.targetLang}} translation for this text.
Do not provide any explanations or text apart from the translation.
{{.sourceLang}}: {{.sourceText}}

{{.targetLang}}:`

	oneChunkReflectionSystemMessage = `You are an expert linguist specializing in translation from {{.sourceLang}} to {{.targetLang}}.
You will be provided with a source text and its translation and your goal is to improve the translation.`
	oneChunkReflectionPrompt = `Your task is to carefully read a source text and a translation from {{.sourceLang}} to {{.targetLang}}, and then give constructive criticisms and helpful suggestions to improve the translation.

The source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as follows:

<SOURCE_TEXT>
{{.sourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.translation1}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's \n
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.targetLang}}).\n

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`
	oneChunkReflectionCountryPrompt = `Your task is to carefully read a source text and a translation from {{.sourceLang}} to {{.targetLang}}, and then give constructive criticism and helpful suggestions to improve the translation.
The final style and tone of the translation should match the style of {{.targetLang}} colloquially spoken in {{.country}}.

The source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as follows:

<SOURCE_TEXT>
{{.sourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.translation1}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's \n
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.targetLang}}).\n

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`

	oneChunkImproveTranslationSystemMessage = `You are an expert linguist, specializing in translation editing from {{.sourceLang}} to {{.targetLang}}.`
	oneChunkImproveTranslationPrompt        = `Your task is to carefully read, then edit, a translation from {{.sourceLang}} to {{.targetLang}}, taking into
account a list of expert suggestions and constructive criticisms.

The source text, the initial translation, and the expert linguist suggestions are delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT>, <TRANSLATION></TRANSLATION> and <EXPERT_SUGGESTIONS></EXPERT_SUGGESTIONS>
as follows:

<SOURCE_TEXT>
{{.sourceText}}
</SOURCE_TEXT>

<TRANSLATION>
{{.translation1}}
</TRANSLATION>

<EXPERT_SUGGESTIONS>
{{.reflection}}
</EXPERT_SUGGESTIONS>

Please take into account the expert suggestions when editing the translation. Edit the translation by ensuring:

(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text)
(iv) terminology (inappropriate for context, inconsistent use), or
(v) other errors.

Output only the new translation and nothing else.`
)

// multi chunk translation
const (
	multiChunkInitialTranslationSystemMessage = `You are an expert linguist, specializing in translation from {{.sourceLang}} to {{.targetLang}}.`
	multiChunkInitialTranslationPrompt        = `Your task is to provide a professional translation from {{.sourceLang}} to {{.targetLang}} of PART of a text.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>. Translate only the part within the source text
delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS>. You can use the rest of the source text as context, but do not translate any
of the other text. Do not output anything other than the translation of the indicated part of the text.

<SOURCE_TEXT>
{{.taggedText}}
</SOURCE_TEXT>

To reiterate, you should translate only this part of the text, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.chunkToTranslate}}
</TRANSLATE_THIS>

Output only the translation of the portion you are asked to translate, and nothing else.`

	multiChunkReflectionSystemMessage = `You are an expert linguist specializing in translation from {{.sourceLang}} to {{.targetLang}}.
You will be provided with a source text and its translation and your goal is to improve the translation.`
	multiChunkReflectionPrompt = `Your task is to carefully read a source text and part of a translation of that text from {{.sourceLang}} to {{.targetLang}}, and then give constructive criticism and helpful suggestions for improving the translation.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context for critiquing the translated part.

<SOURCE_TEXT>
{{.taggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.chunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.translation1Chunk}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's:\n
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.targetLang}}).\n

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`
	multiChunkReflectionCountryPrompt = `Your task is to carefully read a source text and part of a translation of that text from {{.sourceLang}} to {{.targetLang}}, and then give constructive criticism and helpful suggestions for improving the translation.
The final style and tone of the translation should match the style of {{.targetLang}} colloquially spoken in {{.country}}.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context for critiquing the translated part.

<SOURCE_TEXT>
{{.taggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.chunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.translation1Chunk}}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's:\n
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {{.targetLang}}).\n

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else.`

	multiChunkImproveTranslationSystemMessage = `You are an expert linguist, specializing in translation editing from {{.sourceLang}} to {{.targetLang}}.`
	multiChunkImproveTranslationPrompt        = `Your task is to carefully read, then improve, a translation from {{.sourceLang}} to {{.targetLang}}, taking into
account a set of expert suggestions and constructive criticisms. Below, the source text, initial translation, and expert suggestions are provided.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
as context, but need to provide a translation only of the part indicated by <TRANSLATE_THIS> and </TRANSLATE_THIS>.

<SOURCE_TEXT>
{{.taggedText}}
</SOURCE_TEXT>

To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
<TRANSLATE_THIS>
{{.chunkToTranslate}}
</TRANSLATE_THIS>

The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
<TRANSLATION>
{{.translation1Chunk}}
</TRANSLATION>

The expert translations of the indicated part, delimited below by <EXPERT_SUGGESTIONS> and </EXPERT_SUGGESTIONS>, are as follows:
<EXPERT_SUGGESTIONS>
{{.reflectionChunk}}
</EXPERT_SUGGESTIONS>

Taking into account the expert suggestions rewrite the translation to improve it, paying attention
to whether there are ways to improve the translation's

(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {{.targetLang}} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text)
(iv) terminology (inappropriate for context, inconsistent use), or
(v) other errors.

Output only the new translation of the indicated part and nothing else.`
)
