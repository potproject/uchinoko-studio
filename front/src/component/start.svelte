<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import ConfigModal from "./config/general.svelte";
    import ConfigCharacterModal from "./config/character.svelte";
    import ConfigEnvModal from "./config/env.svelte";
    import Character from "./character.svelte";
    import type { CharacterConfig, CharacterConfigList } from "../types/character";
    import type { GeneralConfig } from "../types/general";

    const dispatch = createEventDispatcher();
    let micOk: boolean | undefined = undefined;
    let audioOk = false;
    let wsOk = false;
    let webRtcOk = false;
    let start = false;

    let selectCharacterIndex: number|undefined = undefined;
    let showGeneralConfig = false;
    let showEnvConfig = false;
    let showCharacterConfig: CharacterConfig | undefined = undefined;

    let audioOutputDevices: MediaDeviceInfo[] = [];
    let audioOutputDevicesCharacters: string[] = [];

    let characters: CharacterConfigList = { characters: [] };
    export let general: GeneralConfig;
    let audioElement: HTMLAudioElement;
    let mediaStream: MediaStream;

    onMount(async () => {
        const res = await fetch("/v1/config/characters");
        characters = await res.json();
        for (let i = 0; i < characters.characters.length; i++) {
            for (let j = 0; j < characters.characters[i].voice.length; j++) {
                characters.characters[i].voice[j].behavior = [];
            }
        }
        if (characters.characters.length > 0) {
            selectCharacterIndex = 0;
        }
    });

    let onClick = () => {
        start = true;
        // 音声アンロック
        const audio = new AudioContext();
        const source = audio.createMediaElementSource(audioElement);
        source.connect(audio.destination);
        audioElement.play();

        audioElement.onended = () => {
            if (selectCharacterIndex === undefined) {
                return;
            }
            dispatch("start", {
                audio,
                mediaStream,
                selectCharacter: characters.characters[selectCharacterIndex],
                audioOutputDevicesCharacters,
            });
        };
    };

    const checkMic = async () => {
        try {
            mediaStream = await globalThis.navigator.mediaDevices.getUserMedia({
                audio: true,
            });
            micOk = true;

            const devices = await globalThis.navigator.mediaDevices.enumerateDevices();
            audioOutputDevices = devices.filter((d) => d.kind === "audiooutput");
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
        // @ts-ignore
        webRtcOk = !!(globalThis.RTCPeerConnection || globalThis.mozRTCPeerConnection || globalThis.webkitRTCPeerConnection);
    };

    checkMic();
    checkAudio();
    checkWs();
</script>

<div class="max-h-[90vh] overflow-y-auto">
    <div>
        {#if showGeneralConfig}
            <ConfigModal on:close={() => (showGeneralConfig = false)} data={general} />
        {/if}
        {#if showEnvConfig}
            <ConfigEnvModal on:close={() => (showEnvConfig = false)} />
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
        <audio src="/audio/silent.mp3" preload="auto" class="hidden" bind:this={audioElement}></audio>
        <div
            class="card bg-white shadow-lg rounded-3xl h-auto mx-auto border border-cyan-600 border-opacity-50 border-2 w-96 md:w-128 {start
                ? 'animate-scale-out-horizontal'
                : 'animate-scale-in-hor-center'}"
        >
            <div class="card-header p-4 flex m-2">
                <h1 class="text-3xl font-bold flex-1">Uchinoko Studio(β)</h1>
                <!--<div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => (showGeneralConfig = !showGeneralConfig)}>
                    <i class="las la-cog text-4xl mr-2"></i>
                </div>
                <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => (showEnvConfig = !showEnvConfig)}>
                    <i class="las la-database text-4xl mr-2"></i>
                </div>-->
            </div>
            <!-- 利用規約欄 Textarea -->
            <div class="card-body p-3 m-2">
                <h2 class="text-2xl font-bold text-blue-500">
                    <i class="las la-file-alt"></i>
                    <span>利用規約</span>
                </h2>
                <textarea class="w-full h-32 p-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring focus:ring-blue-500 focus:border-blue-500" placeholder="利用規約" readonly>
    3人のかわいいうちのAIとマイクで会話できるサービスです(大晦日限定公開予定)。

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
                <h2 class="text-2xl font-bold {micOk && audioOk && wsOk && webRtcOk ? 'text-green-600' : 'text-red-600'}">
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
                <div class="flex items-center ml-2 {wsOk && webRtcOk ? 'text-green-500' : 'text-red-500'}">
                    <i class="las la-wifi text-2xl mr-1"></i>
                    <span class="">ネットワーク</span>
                </div>
                {#if !micOk}
                    <div class="text-red-500 text-sm">マイクの権限を「許可する」に設定しないと、音声認識ができません。このサービスを利用するには、マイクの権限を許可してください。</div>
                {/if}
                {#if !audioOk || !wsOk || !webRtcOk}
                    <div class="text-red-500 text-sm">このブラウザでは、このサービスが正常に動作しない可能性があります。推奨するブラウザは、Google ChromeまたはiOS Safariとなります。</div>
                {/if}
                {#if general.characterOutputChange && selectCharacterIndex !== undefined}
                    <div class="border border-gray-300 rounded-md p-2 mt-2">
                        {#each characters.characters[selectCharacterIndex].voice as voice, index}
                            <div class="flex-1 p-2">
                                <label for="voiceOutput" class="text-sm flex items-center"><i class="las la-volume-up text-xl mr-1"></i>{voice.name}</label>
                                <select id="voiceOutput" class="w-full border border-gray-300 rounded p-1" bind:value={audioOutputDevicesCharacters[index]}>
                                    <option value="">システムのデフォルト</option>
                                    {#each audioOutputDevices as device}
                                    <option value={device.deviceId}>{device.label}</option>
                                    {/each}
                                </select>
                            </div>
                        {/each}
                    </div>
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
</div>
