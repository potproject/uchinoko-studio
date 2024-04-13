<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import ConfigModal from "./config/general.svelte";
    import ConfigCharacterModal from "./config/character.svelte";
    import Character from "./character.svelte";
    import type { CharacterConfig, CharacterConfigList } from "../types/character";

    export let id: string;

    const dispatch = createEventDispatcher();
    let micOk: boolean | undefined = undefined;
    let audioOk = false;
    let wsOk = false;
    let start = false;

    let showConfig = false;
    let selectCharacterIndex: number|undefined = undefined;
    let showCharacterConfig: CharacterConfig | undefined = undefined;

    let characters: CharacterConfigList = { characters: [] };


    let slientAudio: HTMLAudioElement;

    let onReset = () => {
        fetch(`/v1/chat/${id}`, {
            method: "DELETE",
        }).finally(() => {
            location.reload();
            localStorage.removeItem("id");
        });
    };

    onMount(async () => {
        const res = await fetch("/v1/config/characters");
        characters = await res.json();
        if (characters.characters.length > 0) {
            selectCharacterIndex = 0;
        }
    });

    let onClick = () => {
        start = true;
        // 音声アンロック
        const audio = new AudioContext();
        const source = audio.createMediaElementSource(slientAudio);
        source.connect(audio.destination);
        slientAudio.play();

        slientAudio.onended = () => {
            if (selectCharacterIndex === undefined) {
                return;
            }
            dispatch("start", {
                audio,
                selectCharacter: characters.characters[selectCharacterIndex],
            });
        };
    };

    const checkMic = async () => {
        try {
            await globalThis.navigator.mediaDevices.getUserMedia({
                audio: true,
            });
            micOk = true;
        } catch (e) {
            micOk = false;
        }
    };

    const checkAudio = () => {
        // @ts-ignore
        audioOk = !!(globalThis.AudioContext || globalThis.webkitAudioContext);
    };

    const checkWs = () => {
        // @ts-ignore
        wsOk = !!(globalThis.WebSocket || globalThis.MozWebSocket);
    };

    checkMic();
    checkAudio();
    checkWs();
</script>

<div>
    {#if showConfig}
        <ConfigModal on:close={() => (showConfig = false)} data={{ transcription: { type: "whisper" } }} />
    {/if}
    {#if showCharacterConfig !== undefined}
        <ConfigCharacterModal
            on:close={() => (showCharacterConfig = undefined)}
            on:update-close={(e) => {
                showCharacterConfig = undefined;
                let newCharacters = characters.characters;
                let exist = false;
                for (let i = 0; i < newCharacters.length; i++) {
                    if (newCharacters[i].general.id === e.detail.general.id) {
                        newCharacters[i] = e.detail;
                        exist = true;
                        break;
                    }
                }
                if (!exist) {
                    newCharacters.push(e.detail);
                }
                characters = { characters: newCharacters };
            }}
            data={showCharacterConfig}
        />
    {/if}
    <audio src="/audio/silent.mp3" preload="auto" class="hidden" bind:this={slientAudio}></audio>
    <div
        class="card bg-white shadow-lg rounded-3xl h-auto mx-auto border border-cyan-600 border-opacity-50 border-2 w-96 md:w-128 {start
            ? 'animate-scale-out-horizontal'
            : 'animate-scale-in-hor-center'}"
    >
        <div class="card-header p-4 flex m-2">
            <h1 class="text-3xl font-bold flex-1">Uchinoko Studio(β)</h1>
            <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => (showConfig = !showConfig)}>
                <i class="las la-cog text-4xl mr-2"></i>
            </div>
        </div>
        <div class="card-header mt-2 px-4 text-xs">
            <div>Chat ID: {id}</div>
            <a href="#" class="text-blue-500 hover:text-blue-600" on:click={onReset}>チャット履歴をリセットする</a>
        </div>
        <!-- 利用規約欄 Textarea -->
        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold text-blue-500">
                <i class="las la-file-alt"></i>
                <span>利用規約</span>
            </h2>
            <textarea class="w-full h-32 p-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring focus:ring-blue-500 focus:border-blue-500" placeholder="利用規約" readonly>
このWebアプリケーションを使用して、起こったいかなる問題についても、開発者は責任を負いません。
音声生成処理に外部のサービスまたはアプリケーションを使用している場合、そのサービスまたはアプリケーションの利用規約に従ってください。 このWebアプリケーションは、Google
Chromeを推奨しています。その他のブラウザでは、正常に動作しない可能性があります。
このアプリケーションはオープンソースであり、ソースコードはGitHubで公開されています。ソースコードの利用は、Githubに記載されているライセンスに従ってください。
https://github.com/potproject/uchinoko-studio</textarea
            >
        </div>

        <!-- 概要欄 -->
        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold text-blue-500">
                <i class="las la-user"></i>
                <span>キャラクター選択</span>
            </h2>
            <div class="flex items-center">
                <Character selectCharacterIndex={selectCharacterIndex} 
                on:selectCharacter={(e) => {
                    selectCharacterIndex = e.detail.index;
                }}
                on:deleteCharacter={(e) => {
                    selectCharacterIndex = undefined;
                    fetch(`/v1/config/character/${e.detail.character.general.id}`, {
                        method: "DELETE",
                    }).then((res) => {
                        if (!res.ok) {
                            throw new Error("削除に失敗しました。");
                        }
                        let newCharacters = characters.characters.filter((c) => c.general.id !== e.detail.character.general.id);
                        characters = { characters: newCharacters };
                    }).catch((e) => {
                        alert(e.message);
                        console.error(e);
                    });
                }}
                on:openCharacterConfig={(e) => {
                    if (e.detail.character === null) {
                        fetch("/v1/config/character/init", {
                            method: "GET",
                        }).then((res) => res.json()).then((data) => {
                            showCharacterConfig = data;
                        });
                        return
                    }
                    showCharacterConfig = e.detail.character
                }} characters={characters} />
            </div>
        </div>

        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold {micOk && audioOk && wsOk ? 'text-green-600' : 'text-red-600'}">
                <i class="las {micOk && audioOk && wsOk ? 'la-check' : 'la-times'}"></i>
                <span>動作チェック</span>
            </h2>
            <div class="flex items-center ml-2 {micOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-microphone text-2xl mr-1"></i>
                <span class="">マイク</span>
            </div>
            <div class="flex items-center ml-2 {audioOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-volume-up text-2xl mr-1"></i>
                <span class="">スピーカー</span>
            </div>
            <div class="flex items-center ml-2 {wsOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-wifi text-2xl mr-1"></i>
                <span class="">ネットワーク</span>
            </div>
            {#if !micOk}
                <div class="text-red-500 text-sm">マイクの権限を「許可する」に設定しないと、音声認識ができません。このサービスを利用するには、マイクの権限を許可してください。</div>
            {/if}
            {#if !audioOk || !wsOk}
                <div class="text-red-500 text-sm">このブラウザでは、このサービスが正常に動作しない可能性があります。推奨するブラウザは、Google ChromeまたはiOS Safariとなります。</div>
            {/if}
            <!-- ボタン -->
            <div class="card-footer p-4 flex m-2 justify-center items-center">
                <button
                    on:click={onClick}
                    disabled={!micOk || !audioOk || !wsOk || selectCharacterIndex === undefined}
                    class="btn bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-full {micOk && audioOk && wsOk && selectCharacterIndex !== undefined ? 'opacity-100' : 'opacity-50 cursor-not-allowed'}"
                    >同意してはじめる</button
                >
            </div>
        </div>
    </div>
</div>
