<script lang="ts">
    import { onMount } from "svelte";

    let timer = 0;
    export let stopMic = false;

    const parseTime = (time: number) => {
        const min = Math.floor(time / 60);
        const sec = time % 60;
        return `${min}:${sec < 10 ? '0' : ''}${sec}`;
    };

    onMount(() => {
        const interval = setInterval(() => {
            if (!stopMic) {
                timer++;
            }
        }, 1000);

        return () => {
            clearInterval(interval);
        };
    });
</script>

<div class="flex justify-center items-center">
    <p class="text-white rounded-md px-2 py-1 m-2 {!stopMic ? 'bg-blue-600' : 'bg-red-600'}">
        {parseTime(timer)}
    </p>
</div>