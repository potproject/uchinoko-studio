<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";
    import { onMount } from "svelte";

    let route: "start" | "chat" = "start";
    let audio: AudioContext;
    let mediaStream: MediaStream;
    let selectCharacter: CharacterConfig;
    let audioOutputDevicesCharacters: string[];
    let general: GeneralConfig = {
        background: "blue",
        language: "ja-JP", 
        soundEffect: true,
        characterOutputChange: false,
        enableTTSOptimization: false,
        transcription: { type: "openai_speech_to_text", method: "auto", autoSetting: { threshold: 0.02, silentThreshold: 1, audioMinLength: 1.3 } },
    };

    let backgroundClass = "from-cyan-100 to-blue-400";

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }

    const onStart = (e: CustomEvent<{ audio: AudioContext, mediaStream: MediaStream ,selectCharacter: CharacterConfig, audioOutputDevicesCharacters: string[] }>) => {
        audio = e.detail.audio;
        mediaStream = e.detail.mediaStream;
        selectCharacter = e.detail.selectCharacter;
        audioOutputDevicesCharacters = e.detail.audioOutputDevicesCharacters;
        changeRoute("chat");
    }

    const setBackGround = (color: string) => {
        if (color === "blue") {
            backgroundClass = "from-cyan-100 to-blue-400";
        } else if (color === "red") {
            backgroundClass = "from-amber-100 to-red-400";
        } else if (color === "green") {
            backgroundClass = "from-lime-100 to-green-400";
        } else if (color === "yellow") {
            backgroundClass = "from-amber-100 to-yellow-400";
        } else if (color === "purple") {
            backgroundClass = "from-violet-100 to-purple-400";
        } else if (color === "pink") {
            backgroundClass = "from-rose-100 to-pink-400";
        } else if (color === "indigo") {
            backgroundClass = "from-blue-100 to-indigo-400";
        } else if (color === "gray") {
            backgroundClass = "from-slate-100 to-gray-400";
        } else if (color === "orange") {
            backgroundClass = "from-red-100 to-orange-400";
        } else if (color === "teal") {
            backgroundClass = "from-emerald-100 to-teal-400";
        }
    }

    onMount(async () => {
        fetch(`/v1/config/general`)
            .then((res) => {
                if (!res.ok) {
                    throw new Error("設定の取得に失敗しました");
                }
                return res.json();
            })
            .then((data: GeneralConfig) => {
                general = data;
                setBackGround(data.background);
            })
            .catch((e) => {
                window.alert(e.message);
                console.error(e);
            });
    });
</script>

<main class={"bg-gradient-to-r w-screen h-screen flex items-center justify-center " + backgroundClass}>
    {#if route === "start"}
    <Start on:start={onStart} general={general} />
    {/if}
    {#if route === "chat"}
    <Chat audio={audio} media={mediaStream} selectCharacter={selectCharacter} audioOutputDevicesCharacters={audioOutputDevicesCharacters} generalConfig={general} />
    {/if}
</main>
