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
        <div class="bg-white shadow-lg rounded-3xl w-128">
            <div class="card-header p-4 flex m-2">
                <h1 class="text-2xl font-bold flex-1">
                    <i class="las la-cog text-2xl mr-2"></i>
                    設定
                </h1>
                <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => dispatch("close")}>
                    <i class="las la-times text-2xl"></i>
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
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="transcription" class="text-sm">Speech to text</label>
                    <select id="transcription" class="w-full border border-gray-300 rounded p-1" bind:value={data.transcription.type}>
                        <option value="openai_speech_to_text">OpenAI Speech to Text API</option>
                        <option value="google_speech_to_text">Google Speech to Text</option>
                        <option value="speech_recognition">SpeechRecognition</option>
                    </select>
                </div>
            </div>

            {#if data.transcription.type === "openai_speech_to_text" || data.transcription.type === "google_speech_to_text"}
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
