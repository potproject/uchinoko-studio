<script lang="ts">
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import { tick } from "svelte";
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

    let initLoading = true;
    let stopMic = false;
    let startScreenCapture = false;

    let socket: SocketContext;
    let playing: PlayingContext;
    let recording: RecordingContentInterface;
    let image: ImageContext;
    let messages: Message[] = [];
    let screenCapture: ScreenCapture = new ScreenCapture();

    export let audio: AudioContext;
    export let media: MediaStream;
    export let selectCharacter: CharacterConfig;
    export let audioOutputDevicesCharacters: string[];
    export let generalConfig: GeneralConfig;
    let backgroundImage: { path: string; characterChange: boolean } = { path: "", characterChange: false };

    const speakDisabled = (disabled: boolean) => {
        if (stopMic || initLoading) {
            disabled = true;
        }
        recording.changeRecordingAllow(!disabled);
    };

    let speaking = false;
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

    image = new ImageContext();
    image.onLoadStart = (file: File) => {
        addMessage({
            type: "my",
            text: MessageConstants.uploadImage,
            img: URL.createObjectURL(file),
            loading: true,
            speaking: false,
            chunk: false,
            voiceIndex: null,
        });
    };
    image.onLoadEnd = (mimeType: string, arrayBuffer: ArrayBuffer) => {
        socket.sendBinary(mimeType, arrayBuffer, "image.png");
    };

    // 画面キャプチャ
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

    const uploadImage = async () => {
        stopMic = true;
        speakDisabled(stopMic);
        image.upload();
    };

    (async () => {
        socket = await SocketContext.connect(getID(), selectCharacter.general.id);
        socket.onClosed = () => {
            addMessage({
                type: "error",
                text: MessageConstants.disconnected,
                loading: false,
                speaking: false,
                chunk: false,
                voiceIndex: null,
            });
        };
        socket.onBinary = (data) => {
            playing.playWAV(data);
            return;
        };

        socket.onFinish = () => {
            // 再生後停止指示
            if (generalConfig.soundEffect) {
                playing.playAudio("audio/ka.mp3");
            }
            playing.sendFinishAction();
        };

        socket.onChatRequest = (text) => {
            changeLastMessage({ text: text, loading: false, speaking: false });
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
                loading: false,
                speaking: false,
                chunk: false,
                voiceIndex: null,
            });
        };

        // Playing 再生
        playing = new PlayingContext(audio);
        playing.onSpeakingStart = () => {
            speaking = true;
        };
        playing.onSpeackingChunkStart = () => {
            while (chunkMessages.length > 0) {
                const chunkMessage = chunkMessages.shift();
                if (!chunkMessage) {
                    return;
                }
                switch (chunkMessage.type) {
                    case "change-character":
                        if (messages[messages.length - 1].chunk && messages[messages.length - 1].type === "your") {
                            changeLastMessage({
                                text: messages[messages.length - 1].text,
                                loading: false,
                                speaking: false,
                                chunk: false,
                            });
                        }
                        const voiceIndex = selectCharacter.voice.findIndex((v) => v.identification === chunkMessage.text);
                        if (generalConfig.characterOutputChange && audioOutputDevicesCharacters.length > 0 && audioOutputDevicesCharacters[voiceIndex]) {
                            playing.changeOutputDevice(audioOutputDevicesCharacters[voiceIndex]);
                        }
                        addMessage({
                            type: "your",
                            text: MessageConstants.empty,
                            loading: true,
                            speaking: true,
                            chunk: true,
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
                        if (messages[messages.length - 1].chunk) {
                            changeLastMessage({
                                text: messages[messages.length - 1].text + chunkMessage.text,
                                loading: true,
                                speaking: true,
                                chunk: true,
                            });
                            return;
                        }
                        addMessage({
                            type: "your",
                            text: chunkMessage.text,
                            loading: true,
                            speaking: true,
                            chunk: true,
                            voiceIndex: 0,
                        });
                        return;
                }
            }
        };

        playing.onSpeakingEnd = () => {
            speaking = false;
            if (messages[messages.length - 1].speaking) {
                changeLastMessage({
                    loading: false,
                    speaking: false,
                    chunk: false,
                });
            }
            speakDisabled(false);
        };

        // Recording 録音
        if (generalConfig.transcription.type === "speech_recognition") {
            recording = new RecognitionContent(media, socket.mimeType, generalConfig);
        } else if (generalConfig.transcription.method === "pushToTalk") {
            recording = new RecordingPushToTalkContext(media, socket.mimeType, generalConfig);
            stopMic = true;
            speakDisabled(true);
        } else {
            recording = new RecordingContext(media, socket.mimeType, generalConfig);
        }
        await recording.init();

        recording.onSpeakingStart = () => {
            addMessage({
                type: "my",
                text: MessageConstants.speakingStart,
                loading: false,
                speaking: true,
                chunk: false,
                voiceIndex: null,
            });

            updateChat();
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
                return;
            }
            changeLastMessage({
                text: MessageConstants.speakingEnd,
                loading: true,
                speaking: false,
                chunk: false,
            });
            speakDisabled(true);
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

        // old message load
        const res = await fetch(`/v1/chat/${getID()}/${selectCharacter.general.id}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
        });
        if (res.status === 204) {
            initLoading = false;
            updateChat();
            return;
        }
        const oldMessages = (await res.json()).Chat as { role: string; content: string }[];
        const newmessages: Message[] = [];
        for (const msg of oldMessages) {
            if (msg.role === "user" || msg.role === "assistant") {
                newmessages.push({
                    type: msg.role === "user" ? "my" : "your",
                    text: msg.content,
                    loading: false,
                    speaking: false,
                    chunk: false,
                    voiceIndex: null,
                });
            }
        }
        messages = newmessages;
        initLoading = false;
        if (selectCharacter.voice.length > 0) {
            backgroundImage = { path: selectCharacter.voice[0].backgroundImagePath, characterChange: false };
        }
        updateChat();
    })();
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
                {#if initLoading}
                    <div class="flex justify-center items-center">
                        <div class="flex justify-center items-center rounded-md bg-gray-600 p-2 m-2 text-white">
                            <i class="las text-2xl animate-spin la-spinner"></i>Loading...
                        </div>
                    </div>
                {/if}
                {#each messages as msg}
                    {#if msg.type === "my"}
                        <ChatMyMsg message={msg.text} image={msg.img} loading={msg.loading} speaking={msg.speaking} />
                    {:else if msg.type === "your"}
                        <ChatYourMsg
                            name={msg.voiceIndex !== null ? selectCharacter.voice[msg.voiceIndex].name : ""}
                            message={msg.text}
                            loading={msg.loading}
                            speaking={msg.speaking}
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
                            class="btn text-white font-bold py-2 px-4 rounded-full
                        {!startScreenCapture ? 'bg-gray-500 hover:bg-gray-600' : 'bg-red-500 hover:bg-red-600'}"
                            on:click={enableScreenCapture}
                        >
                            <i class="las text-2xl la-desktop"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="チャット履歴をリセットする">
                        <button class="btn text-white font-bold py-2 px-4 rounded-full bg-gray-500 hover:bg-gray-600 disabled:opacity-50" disabled={speaking} on:click={refreshChat}>
                            <i class="las text-2xl la-folder-minus"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="画像をアップロード">
                        <button class="btn text-white font-bold py-2 px-4 rounded-full bg-blue-500 hover:bg-blue-600 disabled:opacity-50" disabled={speaking} on:click={uploadImage}>
                            <i class="las text-2xl la-file-image"></i>
                        </button>
                    </Tooltip>
                    <Tooltip text="音声ミュート">
                        <button
                            class="btn text-white font-bold py-2 px-4 rounded-full
                    {!stopMic ? 'bg-blue-500 hover:bg-blue-600' : 'bg-red-500 hover:bg-red-600'}
                    "
                            on:click={() => {
                                stopMic = !stopMic;
                                speakDisabled(stopMic);
                            }}
                        >
                            <i class="las text-2xl {!stopMic ? 'la-microphone' : 'la-microphone-slash'}"></i>
                        </button>
                    </Tooltip>
                </div>
            </div>
        </div>
    </div>
</div>
