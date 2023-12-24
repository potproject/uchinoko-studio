<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";
    import { onMount } from 'svelte';

    let route: "start" | "chat" = "start";
    let audio: AudioContext;
    let selected = "bertvits2";
    
    let id:string;

    onMount(() => {
        const localId = localStorage.getItem("id");
        if (localId) {
            id = localId;
        } else {
            id = self.crypto.randomUUID();
        }
        localStorage.setItem("id", id);
    });


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
    <Start on:start={onStart} id={id} />
    {/if}
    {#if route === "chat"}
    <Chat audio={audio} selected={selected} id={id} />
    {/if}
</main>
