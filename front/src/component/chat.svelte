<script lang="ts">
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import { onMount, tick } from "svelte";
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";
    import ChatError from "./chat-error.svelte";
    import { type Message, type ChunkMessage, MessageConstants } from "../types/message";
    import { RecordingPushToTalkContext } from "$lib/RecordingPushToTalkContent";
    import { RecognitionContent } from "$lib/RecognitionContent";
    import type { RecordingContentInterface } from "$lib/RecordingContentInterface";
    import { ImageContext } from "$lib/ImageContext";
    import { getID } from "$lib/GetId";
    import { ScreenCapture } from "$lib/ScreenCapture";
    import Tooltip from "./tooltip/tooltip.svelte";

    export let audio: AudioContext;
    export let media: MediaStream;
    export let selectCharacter: CharacterConfig;
    export let audioOutputDevicesCharacters: string[];
    export let generalConfig: GeneralConfig;

    let mute = false;
    let onChangeMute = (value: boolean) => {
        mute = value;
        if (mute) {
            recording.changeRecordingAllow(false);
        } else {
            recording.changeRecordingAllow(state === ChatState.Waiting || state === ChatState.UserSpeaking);
        }
    };

    enum ChatState {
        Initializing = "initializing", // 初期化中
        Waiting = "waiting", // ユーザ発話待機中
        UserSpeaking = "user_speaking", // ユーザが音声発信中
        Loading = "loading", // ユーザ発信完了後，AI応答待ち（ロード中）
        AISpeaking = "ai_speaking", // AIが音声発信中
    }
    let state: ChatState = ChatState.Initializing;
    const onChangeState = (newState: ChatState) => {
        if (!mute && ChatState.UserSpeaking !== newState) {
            const disabled = newState !== ChatState.Waiting;
            if (recording) {
                recording.changeRecordingAllow(!disabled);
            }
        }
        state = newState;
    };

    onMount(async () => {
        onChangeState(ChatState.Initializing);
        await loadSocket();
        await loadPlaying();
        await loadRecording();
        await loadMessages();
        await loadImage();
        await loadScreenCapture();
        onChangeState(ChatState.Waiting);
    });

    let socket: SocketContext;
    const loadSocket = async () => {
        socket = await SocketContext.connect(getID(), selectCharacter.general.id);
        socket.onClosed = () => {
            addMessage({
                type: "error",
                text: MessageConstants.disconnected,
                voiceIndex: null,
            });
        };
        socket.onBinary = (data) => {
            playing.playWAV(data);
            return;
        };

        socket.onFinish = () => {
            setTimeout(() => {
                if (generalConfig.soundEffect) {
                    playing.playAudio("audio/ka.mp3");
                }
            }, 300);
            playing.sendFinishAction();
        };

        socket.onChatRequest = (text) => {
            changeLastMessage({ text: text });
        };

        socket.onChatResponseChangeCharacter = (text) => {
            chunkMessages.push({ type: "change-character", text: text });
        };

        socket.onChatResponseChangeBehavior = (imagePath) => {
            chunkMessages.push({ type: "change-behavior", text: imagePath });
        };

        socket.onChatResponseChunk = (text) => {
            chunkMessages.push({ type: "chat", text: text });
        };

        socket.onError = (text) => {
            addMessage({
                type: "error",
                text: text,
                voiceIndex: null,
            });
        };
    };

    let playing: PlayingContext;
    const loadPlaying = async () => {
        playing = new PlayingContext(audio);
        playing.onSpeakingStart = () => {
            onChangeState(ChatState.AISpeaking);
        };
        playing.onSpeackingChunkStart = () => {
            while (chunkMessages.length > 0) {
                const chunkMessage = chunkMessages.shift();
                if (!chunkMessage) {
                    return;
                }
                switch (chunkMessage.type) {
                    case "change-character":
                        const voiceIndex = selectCharacter.voice.findIndex((v) => v.identification === chunkMessage.text);
                        if (generalConfig.characterOutputChange && audioOutputDevicesCharacters.length > 0 && audioOutputDevicesCharacters[voiceIndex]) {
                            playing.changeOutputDevice(audioOutputDevicesCharacters[voiceIndex]);
                        }
                        addMessage({
                            type: "your",
                            text: MessageConstants.empty,
                            voiceIndex,
                        });
                        backgroundImage = { path: "", characterChange: false };
                        tick().then(() => {
                            backgroundImage = {
                                path: selectCharacter.voice[voiceIndex]?.backgroundImagePath ?? "",
                                characterChange: true,
                            };
                        });
                        continue;
                    case "change-behavior":
                        backgroundImage = { path: chunkMessage.text, characterChange: false };
                        continue;
                    case "chat":
                        if (messages[messages.length - 1].type === "your") {
                            changeLastMessage({
                                text: messages[messages.length - 1].text + chunkMessage.text,
                            });
                            return;
                        }
                        addMessage({
                            type: "your",
                            text: chunkMessage.text,
                            voiceIndex: 0,
                        });
                        return;
                }
            }
        };

        playing.onSpeakingEnd = () => {
            onChangeState(ChatState.Waiting);
        };
    };

    let recording: RecordingContentInterface;
    const loadRecording = async () => {
        // Recording 録音
        if (generalConfig.transcription.type === "speech_recognition") {
            recording = new RecognitionContent(media, socket.mimeType, generalConfig);
        } else if (generalConfig.transcription.method === "pushToTalk") {
            // TODO Push To Talkは一旦無し
            // recording = new RecordingPushToTalkContext(media, socket.mimeType, generalConfig);
            addMessage({
                type: "error",
                text: "Push To Talkは未実装です",
                voiceIndex: null,
            });
        } else {
            recording = new RecordingContext(media, socket.mimeType, generalConfig);
        }
        await recording.init();

        recording.onSpeakingStart = () => {
            addMessage({
                type: "my",
                text: MessageConstants.speakingStart,
                voiceIndex: null,
            });

            updateChat();
            onChangeState(ChatState.UserSpeaking);
            return;
        };
        recording.onSpeakingEnd = (ignore) => {
            if (generalConfig.soundEffect) {
                playing.playAudio("audio/pi.mp3");
            }
            // 最後のメッセージを更新
            if (ignore) {
                messages = messages.slice(0, messages.length - 1);
                updateChat();
                onChangeState(ChatState.Waiting);
                return;
            }
            changeLastMessage({
                text: MessageConstants.speakingEnd,
            });
            onChangeState(ChatState.Loading);
            return;
        };
        recording.onDataAvailable = (event) => {
            // screenが有効な場合は、画面を送信
            if (screenCapture.stream?.active) {
                const imageWithSound = async () => {
                    const image = await screenCapture.capture();
                    changeLastMessage({
                        img: URL.createObjectURL(image),
                    });
                    socket.sendBinaries([
                        {
                            contentType: "image/jpeg",
                            data: image,
                            filename: "screen.jpg",
                        },
                        {
                            contentType: "audio/wav",
                            data: event.data,
                            filename: "audio.wav",
                        },
                    ]);
                    return;
                };
                imageWithSound();
                return;
            }

            socket.sendBinary("audio/wav", event.data, "audio.wav");
        };

        recording.onText = (text: string) => {
            socket.sendText(text);
        };
    };

    export function expandMessage(messageText: string, characterConfig: CharacterConfig): Message[] {
        const messages: Message[] = [];
        const text = messageText;

        // 空文字になっている identification は除外する
        const idTags = characterConfig.voice
            .map((voice, idx) => ({
                tag: voice.identification.trim(),
                voiceIndex: idx,
            }))
            .filter((item) => item.tag !== "");

        // ヘルパー: pos 以降で最も早く現れる identification タグを返す
        function getNextTag(pos: number): { index: number; tag: string; voiceIndex: number } | null {
            let next: { index: number; tag: string; voiceIndex: number } | null = null;
            for (const item of idTags) {
                const idx = text.indexOf(item.tag, pos);
                if (idx !== -1 && (next === null || idx < next.index)) {
                    next = { index: idx, tag: item.tag, voiceIndex: item.voiceIndex };
                }
            }
            return next;
        }

        let pos = 0;
        // 最初の identification タグが見つからなければ、全体を1件として扱う
        const firstTag = getNextTag(pos);
        if (!firstTag) {
            messages.push({
                type: "your",
                voiceIndex: null,
                text: text.trim(),
            });
            return messages;
        }

        // タグより前にテキストがある場合は voiceIndex を null とする
        if (pos < firstTag.index) {
            const beforeText = text.substring(pos, firstTag.index).trim();
            if (beforeText) {
                messages.push({
                    type: "your",
                    voiceIndex: null,
                    text: beforeText,
                });
            }
            pos = firstTag.index;
        }

        // メッセージを順次抽出
        while (pos < text.length) {
            const currentTag = getNextTag(pos);
            if (!currentTag) {
                const remaining = text.substring(pos).trim();
                if (remaining && messages.length > 0) {
                    messages.push({
                        type: "your",
                        voiceIndex: messages[messages.length - 1].voiceIndex,
                        text: remaining,
                    });
                }
                break;
            }

            const segmentStart = currentTag.index + currentTag.tag.length;
            const nextTag = getNextTag(segmentStart);
            const segmentEnd = nextTag ? nextTag.index : text.length;
            const segmentText = text.substring(segmentStart, segmentEnd).trim();
            if (segmentText) {
                messages.push({
                    type: "your",
                    voiceIndex: currentTag.voiceIndex,
                    text: segmentText,
                });
            }
            pos = segmentEnd;
        }

        return messages;
    }

    let messages: Message[] = [];
    let loadMessages = async () => {
        const res = await fetch(`/v1/chat/${getID()}/${selectCharacter.general.id}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
        });
        if (res.status === 204) {
            updateChat();
            return;
        }
        const oldMessages = (await res.json()).Chat as { role: string; content: string }[];
        const newmessages: Message[] = [];
        for (const msg of oldMessages) {
            if (msg.role === "user" || msg.role === "assistant") {
                // assistantで複数話者を使用している場合、複数プッシュする
                if (selectCharacter.multiVoice && msg.role === "assistant") {
                    const splitMessages = expandMessage(msg.content, selectCharacter);
                    newmessages.push(...splitMessages);
                    continue;
                }

                newmessages.push({
                    type: msg.role === "user" ? "my" : "your",
                    text: msg.content,
                    voiceIndex: null,
                });
            }
        }
        messages = newmessages;
        if (selectCharacter.voice.length > 0) {
            backgroundImage = { path: selectCharacter.voice[0].backgroundImagePath, characterChange: false };
        }
        updateChat();
    };

    let image: ImageContext;
    const loadImage = async () => {
        image = new ImageContext();
        image.onLoadStart = (file: File) => {
            addMessage({
                type: "my",
                text: MessageConstants.uploadImage,
                img: URL.createObjectURL(file),
                voiceIndex: null,
            });
        };
        image.onLoadEnd = (mimeType: string, arrayBuffer: ArrayBuffer) => {
            socket.sendBinary(mimeType, arrayBuffer, "image.png");
        };
    };
    const uploadImage = async () => {
        state = ChatState.Loading;
        await image.upload();
    };

    let screenCapture: ScreenCapture;
    let startScreenCapture = false;
    const loadScreenCapture = async () => {
        screenCapture = new ScreenCapture();
        screenCapture.onEnded = () => {
            startScreenCapture = false;
        };
    };
    const enableScreenCapture = async () => {
        screenCapture.onEnded = () => {
            startScreenCapture = false;
        };
        if (screenCapture.stream?.active) {
            screenCapture.stopCapture();
            startScreenCapture = false;
            return;
        }
        try {
            await screenCapture.startCapture();
            startScreenCapture = true;
        } catch (e) {
            console.error(e);
        }
        return;
    };

    let backgroundImage: { path: string; characterChange: boolean } = { path: "", characterChange: false };

    let chatarea: HTMLDivElement | undefined = undefined;
    let chunkMessages: ChunkMessage[] = [];

    const updateChat = async () => {
        //スクロールバーを一番下に移動
        await tick();
        chatarea?.scrollTo(0, chatarea.scrollHeight);
    };

    const addMessage = (message: Message) => {
        message.text = message.text.trim();
        messages = [...messages, message];
        updateChat();
    };

    const changeLastMessage = (message: Partial<Message>) => {
        if (message.text !== undefined) {
            message.text = message.text.trim();
        }
        messages = [
            ...messages.slice(0, messages.length - 1),
            {
                ...messages[messages.length - 1],
                ...message,
            },
        ];
        updateChat();
    };

    const refreshChat = async () => {
        if (globalThis.confirm("チャットをリセットしますか？返答が上手くいかない場合に使用してください。")) {
            fetch(`/v1/chat/${getID()}/${selectCharacter.general.id}`, {
                method: "DELETE",
            })
                .finally(() => {
                    messages = [];
                    updateChat();
                })
                .catch((e) => {
                    console.error(e);
                    alert("エラーが発生しました");
                });
        }
    };
</script>

<div class="w-full h-full">
    <div class="flex flex-col md:flex-row md:justify-center justify-end w-full h-full">
        {#if backgroundImage.path !== ""}
            <div class="flex justify-end items-start md:items-end h-full pt-0 absolute md:static z-0 md:w-80">
                <img src={"images/" + backgroundImage.path} alt="avatar" class={"w-full max-h-full " + (backgroundImage.characterChange ? "animate-slide-in-bck-bottom" : "")} />
            </div>
        {/if}
        <div class="flex flex-col z-10 md:w-256">
            <div class="py-2 px-4 h-80 md:h-full overflow-y-scroll hidden-scrollbar" bind:this={chatarea}>
                {#if state === ChatState.Initializing}
                    <div class="flex justify-center items-center">
                        <div class="flex justify-center items-center rounded-md bg-gray-600 p-2 m-2 text-white">
                            <i class="las text-2xl animate-spin la-spinner"></i>Loading...
                        </div>
                    </div>
                {/if}
                {#each messages as msg}
                    {#if msg.type === "my"}
                        <ChatMyMsg message={msg.text} image={msg.img} />
                    {:else if msg.type === "your"}
                        <ChatYourMsg
                            name={msg.voiceIndex !== null ? selectCharacter.voice[msg.voiceIndex].name : ""}
                            message={msg.text}
                            img={msg.voiceIndex === null ? null : "images/" + selectCharacter.voice[msg.voiceIndex].image}
                        />
                    {:else if msg.type === "error"}
                        <ChatError message={msg.text} />
                    {/if}
                {/each}
            </div>
            <div class="py-4">
                <div class="flex justify-center items-center space-x-2">
                    <Tooltip text="画面共有">
                        <button
                            disabled={state !== ChatState.Waiting}
                            class="btn text-white font-bold py-2 px-4 rounded-full disabled:opacity-50
                        {!startScreenCapture ? 'bg-gray-500 hover:bg-gray-600' : 'bg-red-500 hover:bg-red-600'}"
                            on:click={enableScreenCapture}
                        >
                            <i class="las text-2xl la-desktop"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="チャット履歴をリセットする">
                        <button class="btn text-white font-bold py-2 px-4 rounded-full bg-gray-500 hover:bg-gray-600 disabled:opacity-50" disabled={state !== ChatState.Waiting} on:click={refreshChat}>
                            <i class="las text-2xl la-folder-minus"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="画像をアップロード">
                        <button class="btn text-white font-bold py-2 px-4 rounded-full bg-blue-500 hover:bg-blue-600 disabled:opacity-50" disabled={state !== ChatState.Waiting} on:click={uploadImage}>
                            <i class="las text-2xl la-file-image"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="音声ミュート">
                        <button
                            disabled={state === ChatState.UserSpeaking}
                            class="btn text-white font-bold py-2 px-4 rounded-full disabled:opacity-50
                    {!mute ? 'bg-blue-500 hover:bg-blue-600' : 'bg-red-500 hover:bg-red-600'}
                    "
                            on:click={() => {
                                onChangeMute(!mute);
                            }}
                        >
                            <i class="las text-2xl {!mute ? 'la-microphone' : 'la-microphone-slash'}"></i>
                        </button>
                    </Tooltip>
                </div>
            </div>
        </div>
    </div>
</div>
