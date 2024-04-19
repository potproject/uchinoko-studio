<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";
    import type { CharacterConfig } from "../types/character";

    let route: "start" | "chat" = "start";
    let audio: AudioContext;
    let selectCharacter: CharacterConfig;

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }

    const onStart = (e: CustomEvent<{ audio: AudioContext, selectCharacter: CharacterConfig }>) => {
        audio = e.detail.audio;
        selectCharacter = e.detail.selectCharacter;
        changeRoute("chat");
    }
</script>

<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen flex justify-center items-center">
    {#if route === "start"}
    <Start on:start={onStart} />
    {/if}
    {#if route === "chat"}
    <Chat audio={audio} selectCharacter={selectCharacter} />
    {/if}
</main>
