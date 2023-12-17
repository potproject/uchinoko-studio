<script lang="ts">
    import ChatMyMsg from "./chat-my-msg.svelte";
    import ChatYourMsg from "./chat-your-msg.svelte";

    let timer = 0;
    let timerId: number | undefined = undefined;
    let speaking = false;
    let chatarea: HTMLDivElement | undefined = undefined;

    const startTimer = () => {
        timerId = window.setInterval(() => {
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

    const updateChat = () => {
        //スクロールバーを一番下に移動
        chatarea?.scrollTo(0, chatarea.scrollHeight);
    }
</script>

<div>
    <!-- center img circle -->
    <div class="flex justify-center items-center">
        <img src="https://picsum.photos/200" class="rounded-full w-32 h-32 {speaking ? 'animate-pulsate-fwd' : ''}" />
    </div>
    <!-- Timer -->
    <div class="flex justify-center items-center">
        <p class="text-white bg-blue-600 rounded-md px-2 py-1 m-2">
            {parseTime(timer)}
        </p>
    </div>
    <!-- chat area -->
    <div class="w-screen">
        <div class="flex justify-center items-center">
            <div class="w-full md:w-1/2 h-96 overflow-y-scroll hidden-scrollbar" bind:this={chatarea}>
                {#each Array(10) as _}
                <ChatMyMsg />
                <ChatYourMsg />
                {/each}
            </div>
        </div>
    </div>
    <!-- disconnect button -->
    <div class="flex justify-center items-center">
        <button class="btn bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded-full">
            <i class="las la-phone text-2xl rotate-135 mt-1 mx-2"></i>
        </button>
    </div>
</div>