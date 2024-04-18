<script lang="ts">
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import { tick } from "svelte";
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";
    import ChatTimer from "./chat-timer.svelte";
    import type { CharacterConfig } from "../types/character";
    import ChatError from "./chat-error.svelte";
    import type { Message } from "../types/message";

    let initLoading = true;
    let stopMic = false;

    let playing: PlayingContext;
    let recording: RecordingContext;
    let messages: Message[] = [];

    export let audio: AudioContext;
    export let selectCharacter: CharacterConfig;

    const speakDisabled = (disabled: boolean) => {
        if (stopMic || initLoading) {
            disabled = true;
        }
        recording.changeRecordingAllow(!disabled);
    };

    let speaking = false;
    let chatarea: HTMLDivElement | undefined = undefined;

    const updateChat = async () => {
        //スクロールバーを一番下に移動
        await tick();
        chatarea?.scrollTo(0, chatarea.scrollHeight);
    };

    (async () => {
        const { socket, mimeType } = await SocketContext.connect(selectCharacter);
        socket.onClosed = () => {
            messages = [
                ...messages,
                {
                    type: "error",
                    text: "接続が切断されました。再度ページを読み込んでください。",
                    textChunk: [],
                    loading: false,
                    speaking: false,
                    chunk: false,
                    voiceIndex: null,
                },
            ];
            updateChat();
        };
        socket.onBinary = (data) => {
            playing.playWAV(data);
            return;
        };

        socket.onFinish = () => {
            if (messages[messages.length - 1].chunk) {
                messages = [
                    ...messages.slice(0, messages.length - 1),
                    {
                        type: "your",
                        text: messages[messages.length - 1].text.trim(),
                        textChunk: messages[messages.length - 1].textChunk,
                        loading: false,
                        speaking: true,
                        chunk: false,
                        voiceIndex: messages[messages.length - 1].voiceIndex,
                    },
                ];
                updateChat();
            }

            // 再生後停止指示
            playing.sendFinishAction();
        };

        socket.onChatRequest = (text) => {
            messages = [
                ...messages.slice(0, messages.length - 1),
                {
                    type: "my",
                    text: text.trim(),
                    textChunk: [text.trim()],
                    loading: false,
                    speaking: false,
                    chunk: false,
                    voiceIndex: null,
                },
            ];
            updateChat();
        };

        socket.onChatResponseChangeCharacter = (text) => {
            if (messages[messages.length - 1].chunk && messages[messages.length - 1].type === "your") {
                messages = [
                    ...messages.slice(0, messages.length - 1),
                    {
                        type: "your",
                        text: messages[messages.length - 1].text.trim(),
                        textChunk: messages[messages.length - 1].textChunk,
                        loading: false,
                        speaking: false,
                        chunk: false,
                        voiceIndex: messages[messages.length - 1].voiceIndex,
                    },
                ];
            }
            messages = [
                ...messages,
                {
                    type: "your",
                    text: "",
                    textChunk: [],
                    loading: true,
                    speaking: true,
                    chunk: true,
                    voiceIndex: selectCharacter.voice.findIndex((v) => v.identification === text),
                },
            ];
        };

        socket.onChatResponseChunk = (text) => {
            if (messages[messages.length - 1].chunk) {
                messages = [
                    ...messages.slice(0, messages.length - 1),
                    {
                        type: "your",
                        text: (messages[messages.length - 1].text + text).trim(),
                        textChunk: [...messages[messages.length - 1].textChunk, text],
                        loading: true,
                        speaking: true,
                        chunk: true,
                        voiceIndex: messages[messages.length - 1].voiceIndex,
                    },
                ];
                updateChat();
                return;
            }
            messages = [
                ...messages,
                {
                    type: "your",
                    text: text.trim(),
                    textChunk: [text.trim()],
                    loading: true,
                    speaking: true,
                    chunk: true,
                    voiceIndex: 0,
                },
            ];
            updateChat();
            return;
        }

        socket.onError = (text) => {
            messages = [
                ...messages,
                {
                    type: "error",
                    text: text,
                    textChunk: [],
                    loading: false,
                    speaking: false,
                    chunk: false,
                    voiceIndex: null,
                },
            ];
            updateChat();
        };

        // Playing 再生
        playing = new PlayingContext(audio);
        playing.onSpeakingStart = () => {
            speaking = true;
        };
        playing.onSpeakingEnd = () => {
            speaking = false;
            if (messages[messages.length - 1].speaking) {
                messages = [
                    ...messages.slice(0, messages.length - 1),
                    {
                        type: "your",
                        text: messages[messages.length - 1].text,
                        textChunk: messages[messages.length - 1].textChunk,
                        loading: false,
                        speaking: false,
                        chunk: false,
                        voiceIndex: messages[messages.length - 1].voiceIndex,
                    },
                ];
                updateChat();
            }
            speakDisabled(false);
        };

        // Recording 録音
        recording = new RecordingContext(await navigator.mediaDevices.getUserMedia({ audio: true }), mimeType);
        await recording.init();

        recording.onSpeakingStart = () => {
            messages = [
                ...messages,
                {
                    type: "my",
                    text: "...",
                    textChunk: [],
                    loading: false,
                    speaking: true,
                    chunk: false,
                    voiceIndex: null,
                },
            ];
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
            messages = [
                ...messages.slice(0, messages.length - 1),
                {
                    type: "my",
                    text: "Loading...",
                    textChunk: [],
                    loading: true,
                    speaking: false,
                    chunk: false,
                    voiceIndex: null,
                },
            ];
            updateChat();
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
                    textChunk: [msg.content],
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
    <!-- Timer -->
    <ChatTimer {stopMic} />
    <!-- chat area -->
    <div class="w-screen">
        <div class="flex justify-center items-center py-2">
            <div class="w-full md:w-2/3 h-96 overflow-y-scroll hidden-scrollbar" bind:this={chatarea}>
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
    <div class="flex justify-center items-center">
        <div class="flex justify-center items-center">
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
