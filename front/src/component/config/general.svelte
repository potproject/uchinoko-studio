<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { GeneralConfig } from "../../types/general";
    const dispatch = createEventDispatcher();

    export let data: GeneralConfig;

    let saveLoading = false;

    let onSave = (data: GeneralConfig) => {
        saveLoading = true;
        fetch(`/v1/config/general`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => {
                // 200 OK以外のステータスコードが返ってきた場合はエラーとして処理
                if (!res.ok) {
                    throw new Error("設定の保存に失敗しました");
                }
                globalThis.location.reload();
            })
            .catch((e) => {
                window.alert(e.message);
                console.error(e);
            })
            .finally(() => {
                saveLoading = false;
            });
    };
</script>

<div class="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50 py-4">
    <div class="bg-white rounded-lg shadow-lg max-w-lg w-full max-h-full overflow-auto my-10 mx-4">
        <div class="bg-white shadow-lg rounded-3xl w-full">
            <div class="card-header p-4 flex m-2">
                <h1 class="text-2xl font-bold flex-1">
                    <i class="las la-cog text-2xl mr-2"></i>
                    設定
                </h1>
                <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => dispatch("close")}>
                    <i class="las la-times text-2xl"></i>
                </div>
            </div>
            <!-- 表示設定 -->
            <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
                <i class="las la-desktop text-2xl mr-2"></i>
                表示設定
            </h2>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="background" class="text-sm">背景色</label>
                    <div class="flex items-center space-x-2 mt-1">
                        <button class={"w-6 h-6 rounded bg-blue-500 hover:ring-4 ring-offset-2 ring-blue-300 " + (data.background === "blue" ? "ring" : "")} on:click={() => data.background = "blue"}></button>
                        <button class={"w-6 h-6 rounded bg-red-500 hover:ring-4 ring-offset-2 ring-red-300 " + (data.background === "red" ? "ring" : "")} on:click={() => data.background = "red"}></button>
                        <button class={"w-6 h-6 rounded bg-green-500 hover:ring-4 ring-offset-2 ring-green-300 " + (data.background === "green" ? "ring" : "")} on:click={() => data.background = "green"}></button>
                        <button class={"w-6 h-6 rounded bg-yellow-500 hover:ring-4 ring-offset-2 ring-yellow-300 " + (data.background === "yellow" ? "ring" : "")} on:click={() => data.background = "yellow"}></button>
                        <button class={"w-6 h-6 rounded bg-purple-500 hover:ring-4 ring-offset-2 ring-purple-300 " + (data.background === "purple" ? "ring" : "")} on:click={() => data.background = "purple"}></button>
                        <button class={"w-6 h-6 rounded bg-pink-500 hover:ring-4 ring-offset-2 ring-pink-300 " + (data.background === "pink" ? "ring" : "")} on:click={() => data.background = "pink"}></button>
                        <button class={"w-6 h-6 rounded bg-indigo-500 hover:ring-4 ring-offset-2 ring-indigo-300 " + (data.background === "indigo" ? "ring" : "")} on:click={() => data.background = "indigo"}></button>
                        <button class={"w-6 h-6 rounded bg-gray-500 hover:ring-4 ring-offset-2 ring-gray-300 " + (data.background === "gray" ? "ring" : "")} on:click={() => data.background = "gray"}></button>
                        <button class={"w-6 h-6 rounded bg-orange-500 hover:ring-4 ring-offset-2 ring-orange-300 " + (data.background === "orange" ? "ring" : "")} on:click={() => data.background = "orange"}></button>
                        <button class={"w-6 h-6 rounded bg-teal-500 hover:ring-4 ring-offset-2 ring-teal-300 " + (data.background === "teal" ? "ring" : "")} on:click={() => data.background = "teal"}></button>
                    </div>
                </div>
            </div>
            <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
                <i class="las la-microphone text-2xl mr-2"></i>
                入力設定
            </h2>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="language" class="text-sm">言語</label>
                    <select id="language" class="w-full border border-gray-300 rounded p-1" bind:value={data.language}>
                        <option value="ja-JP">日本語</option>
                    </select>
                </div>
            </div>
            <div class="flex items-center px-4 py-2 justify-between">
                <div class="flex-1 items-center justify-between">
                    <input type="checkbox" id="soundEffect" class="mr-2 justify-center items-center" bind:checked={data.soundEffect}>
                    <label for="soundEffect" class="text-sm">効果音を再生する</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2 justify-between">
                <div class="flex-1 items-center justify-between">
                    <input type="checkbox" id="characterOutputChange" class="mr-2 justify-center items-center" bind:checked={data.characterOutputChange}>
                    <label for="characterOutputChange" class="text-sm">キャラクターごとにオーディオ出力を設定する(実験的)</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2 justify-between">
                <div class="flex-1 items-center justify-between">
                    <input type="checkbox" id="enableTTSOptimization" class="mr-2 justify-center items-center" bind:checked={data.enableTTSOptimization}>
                    <label for="enableTTSOptimization" class="text-sm">TTSの処理並列化を有効にする(処理が高速化されます)</label>
                </div>
            </div>
                
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="transcription" class="text-sm">Speech to text</label>
                    <select id="transcription" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.type}>
                        <option value="openai_speech_to_text">OpenAI Speech to Text API</option>
                        <option value="google_speech_to_text">Google Speech to Text</option>
                        <option value="vosk_server">Vosk Server</option>
                        <option value="speech_recognition">SpeechRecognition</option>
                    </select>
                </div>
            </div>

            {#if data.transcription.type !== "speech_recognition"}
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="method" class="text-sm">音声認識方法</label>
                    <select id="method" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.method}>
                        <option value="auto">自動認識</option>
                        <option value="pushToTalk">プッシュトゥトーク</option>
                    </select>
                </div>
            </div>

            {#if data.transcription.method === "auto"}
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="threshold" class="text-sm">自動認識音量の閾値</label>
                    <input type="number" min="0" max="1" step="0.01" id="threshold" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.autoSetting.threshold}>
                </div>
            </div>

            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="silentThreshold" class="text-sm">無音状態の閾値(秒)</label>
                    <input type="number" id="silentThreshold" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.autoSetting.silentThreshold}>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="audioMinLength" class="text-sm">無視する音声の長さ(秒)</label>
                    <input type="number" min="0" max="10" step="0.1" id="audioMinLength" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.autoSetting.audioMinLength}>
                </div>
            </div>
            {/if}
            {/if}

            <!-- 保存/キャンセル -->
            <div class="flex justify-center items-center p-4">
                <button class={"bg-blue-500 text-white rounded-md p-2 w-24 " + (saveLoading ? " opacity-50 cursor-not-allowed" : "")} on:click={() => onSave(data)}>
                    {#if saveLoading}
                        <i class="las la-spinner animate-spin"></i>
                    {:else}
                        保存
                    {/if}
                </button>
                <button class="bg-gray-300 text-gray-800 rounded-md p-2 w-24 ml-2" on:click={() => dispatch("close")}>キャンセル</button>
            </div>
        </div>
    </div>
</div>
