<script lang="ts">
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import { tick } from "svelte";
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";
    import ChatTimer from "./chat-timer.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";
    import ChatError from "./chat-error.svelte";
    import type { Message, ChunkMessage } from "../types/message";
    import { RecordingPushToTalkContext } from "$lib/RecordingPushToTalkContent";

    let initLoading = true;
    let stopMic = false;

    let playing: PlayingContext;
    let recording: RecordingContext | RecordingPushToTalkContext;
    let messages: Message[] = [];

    export let audio: AudioContext;
    export let media: MediaStream;
    export let selectCharacter: CharacterConfig;
    export let generalConfig: GeneralConfig;

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
        messages = [...messages, message];
        updateChat();
    };

    const changeLastMessage = (message: Partial<Message>) => {
        messages = [
            ...messages.slice(0, messages.length - 1),
            {
                ...messages[messages.length - 1],
                ...message,
            },
        ];
        updateChat();
    };

    (async () => {
        const { socket, mimeType } = await SocketContext.connect(generalConfig, selectCharacter);
        socket.onClosed = () => {
            addMessage({
                type: "error",
                text: "接続が切断されました。再度ページを読み込んでください。",
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
            playing.sendFinishAction();
        };

        socket.onChatRequest = (text) => {
            changeLastMessage({ text: text.trim(), loading: false, speaking: false });
        };

        socket.onChatResponseChangeCharacter = (text) => {
            chunkMessages.push({ type: "change-character", text: text });
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
                                text: messages[messages.length - 1].text.trim(),
                                loading: false,
                                speaking: false,
                                chunk: false,
                            });
                        }
                        addMessage({
                            type: "your",
                            text: "",
                            loading: true,
                            speaking: true,
                            chunk: true,
                            voiceIndex: selectCharacter.voice.findIndex((v) => v.identification === chunkMessage.text),
                        });
                        continue;
                    case "chat":
                        if (messages[messages.length - 1].chunk) {
                            changeLastMessage({
                                text: (messages[messages.length - 1].text + chunkMessage.text).trim(),
                                loading: true,
                                speaking: true,
                                chunk: true,
                            });
                            return;
                        }
                        addMessage({
                            type: "your",
                            text: chunkMessage.text.trim(),
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
        if (generalConfig.transcription.method === "pushToTalk") {
            recording = new RecordingPushToTalkContext(media, mimeType);
            stopMic = true;
            speakDisabled(true);
        } else {
            recording = new RecordingContext(media, mimeType);
        }
        await recording.init();

        recording.onSpeakingStart = () => {
            addMessage({
                type: "my",
                text: "話し中...",
                loading: false,
                speaking: true,
                chunk: false,
                voiceIndex: null,
            });

            updateChat();
            return;
        };
        recording.onSpeakingEnd = (ignore) => {
            // 最後のメッセージを更新
            if (ignore) {
                messages = messages.slice(0, messages.length - 1);
                updateChat();
                return;
            }
            changeLastMessage({
                text: "音声認識中...",
                loading: true,
                speaking: false,
                chunk: false,
            });
            speakDisabled(true);
            return;
        };
        recording.onDataAvailable = (event) => {
            socket.sendBinary(event.data);
        };

        // old message load
        const res = await fetch(`/v1/chat/${selectCharacter.general.id}`, {
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
        updateChat();
    })();
</script>

<div>
    <div class="w-screen absolute top-0 left-0 z-10">
        <div class="flex justify-center items-center py-2">
            <div class="w-full md:w-2/3 h-128 px-20 overflow-y-scroll hidden-scrollbar" bind:this={chatarea}>
                {#if initLoading}
                    <div class="flex justify-center items-center">
                        <div class="flex justify-center items-center rounded-md bg-gray-600 p-2 m-2 text-white">
                            <i class="las text-2xl animate-spin la-spinner"></i>Loading...
                        </div>
                    </div>
                {/if}
                {#each messages as msg}
                    {#if msg.type === "my"}
                        <ChatMyMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} />
                    {:else if msg.type === "your"}
                        <ChatYourMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} img={msg.voiceIndex === null ? null : selectCharacter.voice[msg.voiceIndex].image} />
                    {:else if msg.type === "error"}
                        <ChatError message={msg.text} />
                    {/if}
                {/each}
            </div>
        </div>
    </div>
    <div class="flex justify-center items-center z-10 absolute bottom-0 left-0 w-screen pb-16">
        <!-- Timer -->
        <ChatTimer {stopMic} />
        <div class="flex justify-center items-center ml-8">
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
        </div>
    </div>
</div>
