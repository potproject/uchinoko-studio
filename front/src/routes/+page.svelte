<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";

    let route: "start" | "chat" = "start";

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }

    const onStart = (e: CustomEvent<{ audio: AudioContext }>) => {
        console.log(e.detail);
        changeRoute("chat");

        // audio test play https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3
        /*const audio = e.detail.audio;
        const source = audio.createBufferSource();
        fetch("https://interactive-examples.mdn.mozilla.net/media/cc0-audio/t-rex-roar.mp3")
            .then(res => res.arrayBuffer())
            .then(buf => audio.decodeAudioData(buf))
            .then(buf => {
                source.buffer = buf;
                source.connect(audio.destination);
                source.start();
            })
        */
    }
</script>

<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen flex justify-center items-center">
    {#if route === "start"}
    <Start on:start={onStart} />
    {/if}
    {#if route === "chat"}
    <Chat />
    {/if}
</main>
