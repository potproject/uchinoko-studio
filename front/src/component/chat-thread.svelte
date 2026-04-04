<script lang="ts">
    import { createEventDispatcher, onDestroy, onMount, tick } from "svelte";
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";
    import ChatError from "./chat-error.svelte";
    import { type Message, type ChunkMessage, MessageConstants } from "../types/message";
    import { RecognitionContent } from "$lib/RecognitionContent";
    import type { RecordingContentInterface } from "$lib/RecordingContentInterface";
    import { ImageContext } from "$lib/ImageContext";
    import { ScreenCapture } from "$lib/ScreenCapture";
    import Tooltip from "./tooltip/tooltip.svelte";
    import { buildSessionQuery } from "$lib/ChatSession";
    import type { ChatSessionSummary } from "../types/chat";

    export let audio: AudioContext;
    export let media: MediaStream;
    export let ownerId: string;
    export let sessionId: string;
    export let selectCharacter: CharacterConfig;
    export let audioOutputDevicesCharacters: string[];
    export let generalConfig: GeneralConfig;

    const dispatch = createEventDispatcher<{ meta: ChatSessionSummary }>();
    const AUTO_CONVERSATION_PROMPT = "<Continue with a natural next utterance without waiting for new input from the user, while maintaining the flow of the current conversation>";
    const AUTO_CONVERSATION_DELAY_MS = 600;
    const transientTexts = new Set([
        MessageConstants.empty,
        MessageConstants.uploadImage,
        MessageConstants.speakingStart,
        MessageConstants.speakingEnd,
        MessageConstants.disconnected,
    ]);

    type RequestSource = "user" | "auto" | null;

    enum ChatState {
        Initializing = "initializing",
        Waiting = "waiting",
        UserSpeaking = "user_speaking",
        Loading = "loading",
        AISpeaking = "ai_speaking",
    }
    let disposed = false;
    let state: ChatState = ChatState.Initializing;
    let mute = true;
    const syncRecordingAllow = (targetState: ChatState = state) => {
        if (!recording) {
            return;
        }
        recording.changeRecordingAllow(!mute && !autoConversationEnabled && targetState === ChatState.Waiting);
    };
    let onChangeMute = (value: boolean) => {
        if (!recording || state === ChatState.Initializing || state === ChatState.UserSpeaking) {
            return;
        }
        mute = value;
        syncRecordingAllow();
    };
    const onChangeState = (newState: ChatState) => {
        state = newState;
        if (newState !== ChatState.UserSpeaking) {
            syncRecordingAllow(newState);
        }
    };

    let autoConversationEnabled = false;
    let autoConversationTimer: ReturnType<typeof setTimeout> | undefined = undefined;
    let currentRequestSource: RequestSource = null;
    let expectNewAssistantMessage = false;
    let receivedAudioChunkThisTurn = false;

    const clearTransientUserMessage = () => {
        if (messages.length === 0) {
            return;
        }
        const lastMessage = messages[messages.length - 1];
        if (lastMessage.type !== "my") {
            return;
        }
        if (!transientTexts.has(lastMessage.text)) {
            return;
        }
        messages = messages.slice(0, messages.length - 1);
        emitSessionMeta();
        updateChat();
    };

    const clearAutoConversationTimer = () => {
        if (autoConversationTimer) {
            clearTimeout(autoConversationTimer);
            autoConversationTimer = undefined;
        }
    };

    const queueAutoConversationTurn = () => {
        clearAutoConversationTimer();
        autoConversationTimer = setTimeout(() => {
            autoConversationTimer = undefined;
            if (!autoConversationEnabled || state !== ChatState.Waiting || disposed) {
                return;
            }
            currentRequestSource = "auto";
            expectNewAssistantMessage = true;
            receivedAudioChunkThisTurn = false;
            onChangeState(ChatState.Loading);
            socket.sendText(AUTO_CONVERSATION_PROMPT, "auto-conversation");
        }, AUTO_CONVERSATION_DELAY_MS);
    };

    const setAutoConversationEnabled = (value: boolean) => {
        autoConversationEnabled = value;
        if (!value) {
            clearAutoConversationTimer();
            syncRecordingAllow();
            return;
        }
        syncRecordingAllow();
        if (state === ChatState.Waiting) {
            queueAutoConversationTurn();
        }
    };

    onMount(async () => {
        onChangeState(ChatState.Initializing);
        await loadSocket();
        await loadPlaying();
        await loadRecording();
        await loadMessages();
        await loadImage();
        await loadScreenCapture();
        if (!disposed) {
            onChangeState(ChatState.Waiting);
        }
    });

    onDestroy(() => {
        disposed = true;
        clearAutoConversationTimer();
        setAutoConversationEnabled(false);
        socket?.disconnect();
        recording?.dispose();
        playing?.dispose();
        void screenCapture?.stopCapture();
    });

    let socket!: SocketContext;
    const loadSocket = async () => {
        socket = await SocketContext.connect(ownerId, selectCharacter.general.id, sessionId);
        socket.onClosed = () => {
            if (disposed) {
                return;
            }
            setAutoConversationEnabled(false);
            currentRequestSource = null;
            expectNewAssistantMessage = false;
            addMessage({
                type: "error",
                text: MessageConstants.disconnected,
                voiceIndex: null,
            });
        };
        socket.onBinary = (data) => {
            if (disposed) {
                return;
            }
            receivedAudioChunkThisTurn = true;
            playing.playWAV(data);
            return;
        };

        socket.onFinish = () => {
            if (disposed) {
                return;
            }
            setTimeout(() => {
                if (generalConfig.soundEffect) {
                    playing.playAudio("audio/ka.mp3");
                }
            }, 300);
            if (receivedAudioChunkThisTurn || playing.isPlaying()) {
                playing.sendFinishAction();
            }
        };

        socket.onChatRequest = (text) => {
            if (disposed || currentRequestSource !== "user" || messages.length === 0 || messages[messages.length - 1].type !== "my") {
                return;
            }
            changeLastMessage({ text: text });
        };

        socket.onChatIgnored = () => {
            if (disposed) {
                return;
            }
            clearTransientUserMessage();
            currentRequestSource = null;
            expectNewAssistantMessage = false;
            receivedAudioChunkThisTurn = false;
            onChangeState(ChatState.Waiting);
        };

        socket.onChatResponse = () => {
            if (disposed) {
                return;
            }
            currentRequestSource = null;
            if (receivedAudioChunkThisTurn || playing.isPlaying()) {
                return;
            }
            onChangeState(ChatState.Waiting);
            if (autoConversationEnabled) {
                queueAutoConversationTurn();
            }
        };

        socket.onChatResponseChangeCharacter = (text) => {
            if (!disposed) {
                chunkMessages.push({ type: "change-character", text: text });
            }
        };

        socket.onChatResponseChangeBehavior = (imagePath) => {
            if (!disposed) {
                chunkMessages.push({ type: "change-behavior", text: imagePath });
            }
        };

        socket.onChatResponseChunk = (text) => {
            if (!disposed) {
                chunkMessages.push({ type: "chat", text: text });
            }
        };

        socket.onError = (text) => {
            if (disposed) {
                return;
            }
            setAutoConversationEnabled(false);
            currentRequestSource = null;
            expectNewAssistantMessage = false;
            addMessage({
                type: "error",
                text: text,
                voiceIndex: null,
            });
        };
    };

    let playing!: PlayingContext;
    const loadPlaying = async () => {
        playing = new PlayingContext(audio);
        playing.onSpeakingStart = () => {
            if (!disposed) {
                onChangeState(ChatState.AISpeaking);
            }
        };
        playing.onSpeackingChunkStart = () => {
            if (disposed) {
                return;
            }
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
                        expectNewAssistantMessage = false;
                        backgroundImage = { path: "", characterChange: false };
                        tick().then(() => {
                            if (disposed) {
                                return;
                            }
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
                        if (!expectNewAssistantMessage && messages.length > 0 && messages[messages.length - 1].type === "your") {
                            changeLastMessage({
                                text: messages[messages.length - 1].text + chunkMessage.text,
                            });
                            return;
                        }
                        expectNewAssistantMessage = false;
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
            if (disposed) {
                return;
            }
            onChangeState(ChatState.Waiting);
            if (autoConversationEnabled) {
                queueAutoConversationTurn();
            }
        };
    };

    let recording!: RecordingContentInterface;
    const loadRecording = async () => {
        if (generalConfig.transcription.type === "speech_recognition") {
            recording = new RecognitionContent(media, socket.mimeType, generalConfig);
        } else if (generalConfig.transcription.method === "pushToTalk") {
            addMessage({
                type: "error",
                text: "Push To Talkは未実装です",
                voiceIndex: null,
            });
            recording = new RecognitionContent(media, socket.mimeType, generalConfig);
        } else {
            recording = new RecordingContext(media, socket.mimeType, generalConfig);
        }
        await recording.init();
        syncRecordingAllow();

        recording.onSpeakingStart = () => {
            if (disposed) {
                return;
            }
            currentRequestSource = "user";
            expectNewAssistantMessage = true;
            receivedAudioChunkThisTurn = false;
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
            if (disposed) {
                return;
            }
            if (generalConfig.soundEffect) {
                playing.playAudio("audio/pi.mp3");
            }
            if (ignore) {
                messages = messages.slice(0, messages.length - 1);
                emitSessionMeta();
                currentRequestSource = null;
                expectNewAssistantMessage = false;
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
            if (disposed) {
                return;
            }
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
            if (!disposed) {
                socket.sendText(text);
            }
        };
    };

    export function expandMessage(messageText: string, characterConfig: CharacterConfig): Message[] {
        const messages: Message[] = [];
        const text = messageText;

        const idTags = characterConfig.voice
            .map((voice, idx) => ({
                tag: voice.identification.trim(),
                voiceIndex: idx,
            }))
            .filter((item) => item.tag !== "");

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
        const firstTag = getNextTag(pos);
        if (!firstTag) {
            messages.push({
                type: "your",
                voiceIndex: null,
                text: text.trim(),
            });
            return messages;
        }

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

    const buildChatEndpoint = () => `/v1/chat/${ownerId}/${selectCharacter.general.id}${buildSessionQuery(ownerId, sessionId)}`;

    let messages: Message[] = [];
    const emitSessionMeta = () => {
        const stableMessages = messages.filter((message) => {
            if (message.type === "error") {
                return false;
            }
            if (transientTexts.has(message.text)) {
                return false;
            }
            return message.text.trim() !== "";
        });
        const firstUser = stableMessages.find((message) => message.type === "my");
        const lastStable = stableMessages.length > 0 ? stableMessages[stableMessages.length - 1] : null;

        dispatch("meta", {
            sessionId,
            title: firstUser?.text ?? lastStable?.text ?? "新しいチャット",
            preview: lastStable?.text ?? "",
            messageCount: stableMessages.length,
            isDefault: sessionId === ownerId,
        });
    };

    let loadMessages = async () => {
        const res = await fetch(buildChatEndpoint(), {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
        });
        if (res.status === 204) {
            emitSessionMeta();
            updateChat();
            return;
        }
        const oldMessages = (await res.json()).Chat as { role: string; content: string }[];
        const newmessages: Message[] = [];
        for (const msg of oldMessages) {
            if (msg.role === "user" || msg.role === "assistant") {
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
        emitSessionMeta();
        updateChat();
    };

    let image!: ImageContext;
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
        clearAutoConversationTimer();
        currentRequestSource = "user";
        expectNewAssistantMessage = true;
        receivedAudioChunkThisTurn = false;
        onChangeState(ChatState.Loading);
        await image.upload();
    };

    let screenCapture!: ScreenCapture;
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
        await tick();
        chatarea?.scrollTo(0, chatarea.scrollHeight);
    };

    const addMessage = (message: Message) => {
        message.text = message.text.trim();
        messages = [...messages, message];
        emitSessionMeta();
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
        emitSessionMeta();
        updateChat();
    };

    const refreshChat = async () => {
        if (globalThis.confirm("このチャットをリセットしますか？")) {
            fetch(buildChatEndpoint(), {
                method: "DELETE",
            })
                .finally(() => {
                    messages = [];
                    emitSessionMeta();
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
            <div class="flex justify-end items-start md:items-end h-full pt-0 absolute md:static z-0 md:w-80 pointer-events-none">
                <img src={"images/" + backgroundImage.path} alt="avatar" class={"w-full max-h-full " + (backgroundImage.characterChange ? "animate-slide-in-bck-bottom" : "")} />
            </div>
        {/if}
        <div class="flex flex-col z-10 md:w-256 w-full">
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
                            loading={state === ChatState.AISpeaking && msg === messages[messages.length - 1]}
                        />
                    {:else if msg.type === "error"}
                        <ChatError message={msg.text} />
                    {/if}
                {/each}
            </div>
            <div class="py-4">
                <div class="flex justify-center items-center space-x-2">
                    <Tooltip text="自動会話モード">
                        <button
                            disabled={state === ChatState.Initializing || state === ChatState.UserSpeaking}
                            class="btn text-white font-bold py-2 px-4 rounded-full disabled:opacity-50
                        {autoConversationEnabled ? 'bg-emerald-500 hover:bg-emerald-600' : 'bg-slate-500 hover:bg-slate-600'}"
                            on:click={() => {
                                setAutoConversationEnabled(!autoConversationEnabled);
                            }}
                        >
                            <i class="las text-2xl {autoConversationEnabled ? 'la-robot' : 'la-comments'}"></i>
                        </button>
                    </Tooltip>
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
                    <Tooltip text="このチャットをリセットする">
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
                            disabled={state === ChatState.Initializing || state === ChatState.UserSpeaking}
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
