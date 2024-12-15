# Uchinoko Studio (β)
<img width="350" alt="Uchinoko Studio (β)" src="https://github.com/potproject/uchinoko-studio/assets/6498055/55dcb6d2-aff8-4bc3-9d0b-76bef1d88dae.png">

## Example (Ja)

https://github.com/potproject/uchinoko-studio/assets/6498055/42a8e1fc-be28-4844-87ba-53bb6072ab71

Using OpenAI Transcriptions(Whisper), Style-Bert-VITS2, Anthropic Claude 3 Opus.

__Multiple speakers & Real-Time Voice Conversation.__

WARNING: This repository does not contain character images, LLM and Audio Model Data.

## About

__Work In Progress__

Uchinoko Studio is a web application designed to facilitate real-time voice conversations with AI. This is achieved by integrating various Large Language Models (LLMs), including Speech-To-Text LLMs like [Whisper](https://github.com/openai/whisper), Chat-based LLMs such as [GPT-4](https://openai.com/), and Text-To-Speech LLMs like [Bert-Vits2](https://github.com/fishaudio/Bert-VITS2).

`Uchinoko` is a Japanese word that means `Waifu / My Daughter`.

See Article(Japanese Only): https://blog.potproject.net/2023/12/24/ai-web-uchinoko-studio/

## Features

* Real-time Voice Conversation
* Multiple Speakers (Experimental)
* Image Support (Experimental)
* __Fast Response(Maybe 1 second or less)__
* [Tailscale](https://tailscale.com/) Support
* Run on Browser(Google Chrome Supported)
* Mobile Browser Support(iOS Safari, Android Chrome)
* Japanese Support(for now...)
* Chat-based LLM Support: [OpenAI GPT(Cloud)](https://openai.com/gpt-4),[Anthropic Claude(Cloud)](https://www.anthropic.com/claude),[Cohere Command(Cloud)](https://cohere.com/command),[Gemini(Cloud)](https://gemini.google.com),Local LLM(OpenAI `/v1/chat/completions` Compatible)
* Speach-To-Text Support: [OpenAI Speech to Text API(Cloud)](https://platform.openai.com/docs/guides/speech-to-text),[Google Speech to Text API(Cloud)](https://cloud.google.com/speech-to-text),[Vosk Server(local)](https://github.com/alphacep/vosk-server),[SpeechRecognition(Web API)](https://developer.mozilla.org/docs/Web/API/SpeechRecognition)
* Text-To-Speech Support: [Bert-Vits2(local)](https://github.com/fishaudio/Bert-VITS2), [Style-Bert-VITS2(local)](https://github.com/litagin02/Style-Bert-VITS2), [VOICEVOX(local)](https://voicevox.hiroshiba.jp/),[NijiVoice API(Cloud)](https://nijivoice.com), [Google Text To Speech API(Cloud)](https://cloud.google.com/text-to-speech),  [OpenAI Speech API(Cloud)](https://platform.openai.com/docs/guides/speech-to-text)
* More bugs...

## Getting Started

__!WARNING! このアプリケーションはまだ開発中であり、正常に動作しない場合があります。__
上手く動作しなくなった場合は、databaseフォルダを削除すると初期化されます。

また、アンチウイルスソフトウェアによっては、誤検知され、不正なファイルと判定されることがあるため、その場合は除外設定を行ってください。

[Env Setting](#env-setting)を参考に.env.exampleから.envファイルを作成すれば実行できます。

## Env Setting

### Speech To Text

* OpenAI Speech to Text API(Cloud): `OPENAI_SPEECH_TO_TEXT_API_KEY`
* Google Speech to Text API(Cloud): `GOOGLE_SPEECH_TO_TEXT_API_KEY`
* Vosk Server: `VOSK_SERVER_ENDPOINT`
* SpeechRecognition(Web API): none

### Chat-based LLM(Cloud)

* OpenAI: `OPENAI_API_KEY`
* Anthropic: `ANTHROPIC_API_KEY`
* cohere: `COHERE_API_KEY`
* Gemini: `GEMINI_API_KEY`

### Chat-based LLM(Local)

LM StudioやOllamaのような、OpenAI互換のローカルLLMを使用することが出来ます。

* `OPENAI_LOCAL_API_KEY` and `OPENAI_LOCAL_API_ENDPOINT`

```
# Example (Using LM Studio)
OPENAI_LOCAL_API_KEY=
OPENAI_LOCAL_API_ENDPOINT=http://localhost:1234/
```

```
# Example (Using Ollama)
OPENAI_LOCAL_API_KEY=
OPENAI_LOCAL_API_ENDPOINT=http://localhost:11434/
```

### Text To Speech

このアプリケーションを使用する場合、以下のソフトウェアをローカルまたはネットワーク上で動作させておくことが前提です。

* VOICEVOXの場合: `VOICEVOX_ENDPOINT`にVOICEVOX Engine APIのエンドポイントを設定してください。
* BERTVITS2の場合: `BERTVITS2_ENDPOINT`にBert-VITS2 FastAPIのエンドポイントを設定してください。また、先にモデルのロードを行っていないと動作しません。
* STYLEBERTVIT2の場合: `STYLEBERTVIT2_ENDPOINT`にStyle-Bert-VITS2 API Serverのエンドポイントを設定してください。モデルのロードは自動で行ってくれるため不要です。Bert-VITS2のAPIとの互換性はありません。
* NijiVoice API(Cloud)の場合: `NIJIVOICE_API_KEY`を設定してください。
* Google Text To Speech API(Cloud)の場合: `GOOGLE_TEXT_TO_SPEECH_API_KEY`を設定してください。
* OpenAI Speech API(Cloud)の場合: `OPENAI_SPEECH_API_KEY`を設定してください。

### Tailscale 

* Tailscaleを使用する場合、起動時にコンソールより認証URLが表示されるので、そこから認証を行ってください。Tailscaleのアカウントが必要です。
* `TAILSCALE_ENABLED`を`true`に設定すると、[Tailscale](https://tailscale.com/)を使用してVPN上からアクセスできるようになります。
  * これにより、自宅で起動して外からhttps通信で無いと動作しないSafariやiOSからもアクセスできるようになります。
* `TAILSCALE_FUNNEL_ENABLED`を`true`に設定すると、
[Tailscale Funnel](https://tailscale.com/kb/1223/funnel)機能を使用してパブリックアクセスできるようになります。何のことかわからなければ変更しないでください。

## Development

### Requirements

* Go (Tested on 1.22.2/win-amd64)
* Node.js (Tested on 20.11.1/win-amd64)
* pnpm

### Run on Local

```
## Easy Start (Windows)
git clone https://github.com/potproject/uchinoko-studio
cd uchinoko-studio
copy server\.env.example server\.env
run-win.bat

## Easy Start (Linux/mac)
git clone https://github.com/potproject/uchinoko-studio
cd uchinoko-studio
cp server/.env.example server/.env
run.sh
```

```bash
# Install Dependencies
cd front
pnpm install
pnpm build

# Running
cd ../server
go run main.go

# 自動でブラウザが立ち上がります。
# 立ち上がらない場合は、http://localhost:15000/ にアクセスしてください。
# 話者(チャットプロンプト/使用するモデルの設定)などは、ブラウザより設定が可能です
```

## TODO

* Docs
* Frontend Design issue
* Add Multilingual Support(EN and CH)
* Fix Bugs...


