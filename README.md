# Uchinoko Studio (β)
<img width="350" alt="Uchinoko Studio (β)" src="https://github.com/potproject/uchinoko-studio/assets/6498055/55dcb6d2-aff8-4bc3-9d0b-76bef1d88dae.png">

## Example (Ja)

https://github.com/potproject/uchinoko-studio/assets/6498055/e3b6691f-598a-4d01-b631-f7b64b7602f7

Using OpenAI Transcriptions(Whisper), Style-Bert-VITS2, Anthropic Claude 3 Opus.

__Multiple speakers & Real-Time Voice Conversation.__

## About

__Work In Progress__

Uchinoko Studio is a web application designed to facilitate real-time voice conversations with AI. This is achieved by integrating various Large Language Models (LLMs), including Speech-To-Text LLMs like [Whisper](https://github.com/openai/whisper), Chat-based LLMs such as [GPT-4](https://openai.com/), and Text-To-Speech LLMs like [Bert-Vits2](https://github.com/fishaudio/Bert-VITS2).

`Uchinoko` is a Japanese word that means `Waifu / My Daughter`.

See Article(Japanese Only): https://blog.potproject.net/2023/12/24/ai-web-uchinoko-studio/

## Features

* Real-time Voice Conversation
* Multiple speakers (Experimental)
* __Fast Response(Maybe 1 second or less)__
* [Tailscale](https://tailscale.com/) Support
* Run on Browser(Google Chrome Supported)
* Japanese Support(for now...)
* Chat-based LLM Support: [OpenAI GPT(Cloud Only)](https://openai.com/gpt-4)、[Anthropic Claude](https://www.anthropic.com/claude)、[Cohere Command(Cloud Only)](https://cohere.com/command)、Local LLM(OpenAI `/v1/chat/completions` Compatible)
* STT LLM Support: [Whisper(Cloud Only)](https://openai.com/research/whisper)
* TTS LLM Support: [Bert-Vits2](https://github.com/fishaudio/Bert-VITS2), [Style-Bert-VITS2](https://github.com/litagin02/Style-Bert-VITS2), [VOICEVOX](https://voicevox.hiroshiba.jp/)
* More bugs...

## Getting Started

TODO: 環境不要で動作できるパッケージを配布することを予定しています。現在は以下の環境での動作が必要です。

### Requirements

* Go (Tested on 1.22.2/win-amd64)
* Node.js (Tested on 20.11.1/win-amd64)
* pnpm
* (When using) OpenAI `/v1/chat/completions` Compatible Local LLM (Tested on [LM Studio](https://lmstudio.ai/) - Llama 3 8B)

### Env Setting Up

[.env.example](server/.env.example)を参考に`server/.env`を作成し
てください。

#### Speech To Text And Chat-based LLM

* `OPENAI_API_KEY`は動作に必須です。設定してください。

* `ANTHROPIC_API_KEY`、`COHERE_API_KEY`、`VOICEVOX_ENDPOINT`、`BERTVITS2_ENDPOINT`、`STYLEBERTVIT2_ENDPOINT`は使用するのであれば設定してください。

* `OPENAI_LOCAL_API_KEY`および`OPENAI_LOCAL_API_ENDPOINT`はOpenAI互換エンドポイントを利用したローカルLLMを使用する場合に設定してください。

#### Text To Speech

このアプリケーションを使用する場合、以下のソフトウェアをローカルまたはネットワーク上で動作させておくことが前提です。

* VOICEVOXの場合: `VOICEVOX_ENDPOINT`にVOICEVOX Engine APIのエンドポイントを設定してください。
* BERTVITS2の場合: `BERTVITS2_ENDPOINT`にBert-VITS2 FastAPIのエンドポイントを設定してください。また、先にモデルのロードを行っていないと動作しません。
* STYLEBERTVIT2の場合: `STYLEBERTVIT2_ENDPOINT`にStyle-Bert-VITS2 API Serverのエンドポイントを設定してください。モデルのロードは自動で行ってくれるため不要です。Bert-VITS2のAPIとの互換性はありません。

#### Tailscale 

* Tailscaleを使用する場合、起動時にコンソールより認証URLが表示されるので、そこから認証を行ってください。Tailscaleのアカウントが必要です。
* `TAILSCALE_ENABLED`を`true`に設定すると、[Tailscale](https://tailscale.com/)を使用してVPN上からアクセスできるようになります。
  * これにより、自宅で起動して外からhttps通信で無いと動作しないSafariやiOSからもアクセスできるようになります。
* `TAILSCALE_FUNNEL_ENABLED`を`true`に設定すると、
[Tailscale Funnel](https://tailscale.com/kb/1223/funnel)機能を使用してパブリックアクセスできるようになります。何のことかわからなければ変更しないでください。

### Run on Local

```
## Easy Start (Windows)
run-win.bat

## Easy Start (Linux/mac)
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
* Mobile Browser Support(iOS Safari, Android Chrome)
* Add Multilingual Support(EN and CH)
* Fix Bugs...


