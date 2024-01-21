<script lang="ts">
    import { PlayingContext } from "$lib/PlayingContext";
    import { RecordingContext } from "$lib/RecordingContent";
    import { SocketContext } from "$lib/SocketContext";
    import { tick } from "svelte";
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";

    let initLoading = true;
    let stopMic = false;

    const speakDisabled = (disabled: boolean) => {
        if (stopMic || initLoading) {
            disabled = true;
        }
        recording.changeRecordingAllow(!disabled);
    }

    export let audio: AudioContext;
    export let selected: string;
    export let id: string;

    let timer = 0;
    let timerId: number | undefined = undefined;
    let speaking = false;
    let chatarea: HTMLDivElement | undefined = undefined;


    type Message = {
        type: 'my' | 'your' | 'error';
        text: string;
        loading: boolean;
        speaking: boolean;
        chunk: boolean;
    }
    let messages: Message[] = [];

    const startTimer = () => {
        timerId = window.setInterval(() => {
            if (stopMic) {
                return;
            }
            timer++;
        }, 1000);
    }

    const parseTime = (time: number) => {
        // 00:00
        const min = Math.floor(time / 60);
        const sec = time % 60;
        return `${min < 10 ? '0' + min : min}:${sec < 10 ? '0' + sec : sec}`;
    }
    startTimer();

    const updateChat = async () => {
        //スクロールバーを一番下に移動
        await tick();
        chatarea?.scrollTo(0, chatarea.scrollHeight);
    }

    let socket: SocketContext;
    let playing: PlayingContext;
    let recording: RecordingContext;

    (async () => {
        // WS
        const wsTLS = location.protocol === 'https:' ? 'wss' : 'ws';

        const extenstion = MediaRecorder.isTypeSupported('audio/webm') ? 'webm' : 'mp4';
        const mimeType = `audio/${extenstion}`;

        const url = `${wsTLS}://${location.host}/v1/ws/talk/${id}/${selected}/${extenstion}`;
        socket = new SocketContext(url);
        await new Promise(resolve => {
            socket.onConnected = () => {
                resolve(null);
            }
        });
        socket.onClosed = () => {
            messages = [...messages, {
                type: 'error',
                text: '接続が切断されました。再度ページを読み込んでください。',
                loading: false,
                speaking: false,
                chunk: false
            }];
            updateChat();
        }
        socket.onBinary = (data) => {
            if (selected === 'bertvits2' || selected === 'voicevox') {
                playing.playWAV(data);
                return;
            }
            playing.playPCM(data);
        }
        socket.onText = (data) => {
            if(data.type === 'finish') {
                if (messages[messages.length - 1].chunk) {
                    messages = [...messages.slice(0, messages.length - 1), {
                        type: 'your',
                        text: messages[messages.length - 1].text,
                        loading: false,
                        speaking: true,
                        chunk: false
                    }];
                    updateChat();
                }

                // 再生後停止指示
                playing.sendFinishAction();
                return;
            }
            if(data.type === 'chat-request') {
                messages = [...messages.slice(0, messages.length - 1), {
                    type: 'my',
                    text: data.text,
                    loading: false,
                    speaking: false,
                    chunk: false
                }];
                updateChat();
                return;
            }

            if(data.type === 'chat-response-chunk') {
                if (messages[messages.length - 1].chunk) {
                    messages = [...messages.slice(0, messages.length - 1), {
                        type: 'your',
                        text: messages[messages.length - 1].text + data.text,
                        loading: true,
                        speaking: true,
                        chunk: true
                    }];
                    updateChat();
                    return;
                }
                messages = [...messages, {
                    type: 'your',
                    text: data.text,
                    loading: true,
                    speaking: true,
                    chunk: true
                }];
                updateChat();
                return;
            }

            /*if(data.type === 'chat-response') {
                if (messages[messages.length - 1].chunk) {
                    messages = [...messages.slice(0, messages.length - 1), {
                        type: 'your',
                        text: data.text,
                        loading: false,
                        speaking: false,
                        chunk: false
                    }];
                    updateChat();
                    return;
                }
                messages = [...messages, {
                    type: 'your',
                    text: data.text,
                    loading: false,
                    speaking: false,
                    chunk: false
                }];
                return;
            }*/

            if(data.type === 'error') {
                messages = [...messages, {
                    type: 'error',
                    text: data.text,
                    loading: false,
                    speaking: false,
                    chunk: false
                }];
                updateChat();
                return;
            }
        }

        // Playing 再生
        playing = new PlayingContext(audio);
        playing.onSpeakingStart = () => {
            speaking = true;
        }
        playing.onSpeakingEnd = () => {
            speaking = false;
            if (messages[messages.length - 1].speaking) {
                messages = [...messages.slice(0, messages.length - 1), {
                    type: 'your',
                    text: messages[messages.length - 1].text,
                    loading: false,
                    speaking: false,
                    chunk: false
                }];
                updateChat();
            }
            speakDisabled(false);
        }

        // Recording 録音
        recording = new RecordingContext(await navigator.mediaDevices.getUserMedia({ audio: true }), mimeType);
        await recording.init();

        recording.onSpeakingStart = () => {
            messages = [...messages, {
                type: 'my',
                text: '...',
                loading: false,
                speaking: true,
                chunk: false
            }];
            updateChat();
            return;
        }
        recording.onSpeakingEnd = (ignore) => {
            // 最後のメッセージを更新
            if(ignore) {
                messages = messages.slice(0, messages.length - 1);
                updateChat();
                return;
            }
            messages = [...messages.slice(0, messages.length - 1), {
                type: 'my',
                text: 'Loading...',
                loading: true,
                speaking: false,
                chunk: false
            }];
            updateChat();
            speakDisabled(true);
            return;
        }
        recording.onDataAvailable = (event) => {
            socket.sendBinary(event.data);
        }

        // old message load
        const res = await fetch(`/v1/chat/${id}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if(res.status === 204) {
            initLoading = false;
            updateChat();
            return;
        }
        const oldMessages = (await res.json()).Chat as { role: string, content: string }[];
        const newmessages: Message[] = [];
        for(const msg of oldMessages) {
            if(msg.role === 'user' || msg.role === 'assistant') {
                newmessages.push({
                    type: msg.role === 'user' ? 'my' : 'your',
                    text: msg.content,
                    loading: false,
                    speaking: false,
                    chunk: false
                });
            }
        }
        messages = newmessages;
        initLoading = false;
        updateChat();
    })();
</script>

<div>
    <!-- center img circle -->
    <div class="flex justify-center items-center">
        <img src="default.png" class="rounded-full w-32 h-32 {speaking ? 'animate-pulsate-fwd border-4 border-blue-500' : ''}" alt="ai" />
    </div>
    <!-- Timer -->
    <div class="flex justify-center items-center">
        <p class="text-white rounded-md px-2 py-1 m-2 {!stopMic ? 'bg-blue-600' : 'bg-red-600'}">
            {parseTime(timer)}
        </p>
    </div>
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
                    {#if msg.type === 'my'}
                        <ChatMyMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} />
                    {:else if msg.type === 'your'}
                        <ChatYourMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} />
                    {:else if msg.type === 'error'}
                    <div class="flex justify-center items-center rounded-md bg-red-600 p-2 m-2 text-white">
                        <i class="las text-2xl la-exclamation-circle"></i>{msg.text}
                    </div>
                    {/if}
                {/each}
            </div>
        </div>
    </div>
    <div class="flex justify-center items-center">
        <div class="flex justify-center items-center">
            <button class="btn text-white font-bold py-2 px-4 rounded-full
            {!stopMic ? 'bg-blue-500 hover:bg-blue-600' : 'bg-red-500 hover:bg-red-600'}
            " on:click={() => {
                stopMic = !stopMic;
                speakDisabled(stopMic);
            }}>
                <i class="las text-2xl {!stopMic ? 'la-microphone' : 'la-microphone-slash'}"></i>
            </button>
        </div>
    </div>
</div>