<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";

    let route: "start" | "chat" = "start";
    let audio: AudioContext;
    let mediaStream: MediaStream;
    let selectCharacter: CharacterConfig;
    let general: GeneralConfig;

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }

    const onStart = (e: CustomEvent<{ audio: AudioContext, mediaStream: MediaStream ,selectCharacter: CharacterConfig, general: GeneralConfig }>) => {
        audio = e.detail.audio;
        mediaStream = e.detail.mediaStream;
        selectCharacter = e.detail.selectCharacter;
        general = e.detail.general;
        changeRoute("chat");
    }
</script>

<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen">
    {#if route === "start"}
    <Start on:start={onStart} />
    {/if}
    {#if route === "chat"}
    <Chat audio={audio} media={mediaStream} selectCharacter={selectCharacter} generalConfig={general} />
    {/if}
</main>
