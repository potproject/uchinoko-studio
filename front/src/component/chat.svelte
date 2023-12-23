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
        if (stopMic){
            disabled = true;
        }
        recording.changeRecordingAllow(!disabled);
    }

    export let audio: AudioContext;

    let timer = 0;
    let timerId: number | undefined = undefined;
    let speaking = false;
    let chatarea: HTMLDivElement | undefined = undefined;


    type Message = {
        type: 'my' | 'your';
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
        const url = `ws://localhost:3000/v1/ws/talk/1/bertvits2/webm`;
        socket = new SocketContext(url);
        await new Promise(resolve => {
            socket.onConnected = () => {
                resolve(null);
            }
        });
        socket.onBinary = (data) => {
            playing.playWAV(data);
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
        recording = new RecordingContext(await navigator.mediaDevices.getUserMedia({ audio: true }));
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
        initLoading = false;
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
        <div class="flex justify-center items-center">
            <div class="w-full md:w-1/2 h-96 overflow-y-scroll hidden-scrollbar" bind:this={chatarea}>
                {#each messages as msg}
                    {#if msg.type === 'my'}
                        <ChatMyMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} />
                    {:else}
                        <ChatYourMsg message={msg.text} loading={msg.loading} speaking={msg.speaking} />
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