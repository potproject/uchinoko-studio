<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";

    let route: "start" | "chat" = "start";
    let audio: AudioContext;
    let selected = "bertvits2";

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }

    const onStart = (e: CustomEvent<{ audio: AudioContext, selected: string }>) => {
        audio = e.detail.audio;
        selected = e.detail.selected;
        changeRoute("chat");
    }
</script>

<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen flex justify-center items-center">
    {#if route === "start"}
    <Start on:start={onStart} />
    {/if}
    {#if route === "chat"}
    <Chat audio={audio} selected={selected} />
    {/if}
</main>
