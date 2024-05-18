<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { CharacterConfig } from "../../types/character";
    const dispatch = createEventDispatcher();

    let showVoice = false;
    let showChat = false;

    let saveLoading = false;

    export let data: CharacterConfig;

    const uploadImage = (index: number) => {
        // 画像アップロードの処理
        // ファイル選択
        const input = document.createElement("input");
        input.type = "file";
        input.accept = "image/*";
        // ファイル選択時の処理
        input.onchange = async (e) => {
            // @ts-ignore
            const file = input.files[0];
            const reader = new FileReader();
            reader.onload = (e) => {
                // base64エンコードした画像をdataURLとして取得
                data.voice[index].image = reader.result as string;
            };
            reader.readAsDataURL(file);
        };
        input.click();
    };

    const onReset = () => {
        if (window.confirm("チャット履歴をリセットしますか？")) {
            fetch(`/v1/chat/${data.general.id}`, {
                method: "DELETE",
            }).finally(() => {
                location.reload();
            });
            alert("チャット履歴をリセットしました");
        }
    };

    const onSave = () => {
        saveLoading = true;
        fetch(`/v1/config/character/${data.general.id}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        })
            .then((res) => {
                if (!res.ok) {
                    throw new Error("設定の保存に失敗しました");
                }
                dispatch("update-close", data);
            })
            .catch((e) => {
                window.alert(e.message);
                console.error(e);
            }).finally(() => {
                saveLoading = false;
            });
    };
</script>

<div class="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50 py-4">
    <div class="bg-white rounded-lg shadow-lg max-w-lg w-full max-h-full overflow-auto my-10 mx-4">
        <div class="card-header p-4 flex m-2">
            <h1 class="text-2xl font-bold flex-1">
                <i class="las la-wrench text-2xl mr-2"></i>
                キャラクター設定
            </h1>
            <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => dispatch("close")}>
                <i class="las la-times text-2xl"></i>
            </div>
        </div>
        <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
            <i class="las la-user text-2xl mr-2"></i>
            基本設定
        </h2>
        <div class="flex items-center px-4 py-2">
            <div class="flex-1">
                <label for="name" class="text-sm">ID</label>
                <div class="flex items-center">
                    <input type="text" id="name" class="w-full border border-gray-300 rounded p-1" readonly disabled bind:value={data.general.id} />
                </div>
            </div>
        </div>
        <div class="flex items-center px-4 py-2">
            <!-- 名前 -->
            <div class="flex-1">
                <label for="name" class="text-sm">キャラクター名</label>
                <input type="text" id="name" class="w-full border border-gray-300 rounded p-1" bind:value={data.general.name} />
            </div>
        </div>

        <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
            <i class="las la-microphone text-2xl mr-2"></i>
            音声設定
            <i class={"las text-2xl ml-auto" + (showVoice ? " la-angle-up" : " la-angle-down")} on:click={() => (showVoice = !showVoice)}></i>
        </h2>
        {#if showVoice}
            <!-- 複数音声を有効化 -->
            <div class="flex items-center px-4 py-2 justify-between">
                <div class="flex-1 items-center justify-between">
                    <input type="checkbox" id="multi_voice" class="mr-2" bind:checked={data.multiVoice} on:change={
                        () => {
                            // 複数音声を無効化した場合、最初の音声設定のみ残す
                            if(!data.multiVoice){
                                data.voice = [data.voice[0]];
                            }
                        }
                    } />
                    <label for="multi_voice" class="text-sm">複数音声を有効化(実験的機能)</label>
                </div>
            </div>  
            {#each data.voice as _, index}
                <div class="border border-gray-300 rounded p-2 m-2">
                    <div class="flex items-center px-4 py-2">
                        <!-- キャラクター画像 -->
                        <div class="flex-1">
                            <label for="image" class="text-sm">キャラクター画像</label>
                            <img src={data.voice[index].image} alt="キャラクター画像" class="w-24 h-24 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2" on:click={() => uploadImage(index)} />
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="backgroundImage" class="text-sm">立ち絵画像</label>
                            <input type="text" id="backgroundImage" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].backgroundImagePath} placeholder="画像パス" />
                            <p class="text-xs text-gray-500">※ imagesフォルダ内に画像を配置してください</p>
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="behavior" class="text-sm">立ち絵変更</label>
                            <div class="border">
                                {#each data.voice[index].behavior as _, i}
                                <div class="bg-white p-2 rounded-lg w-full flex border-gray-300">
                                    <button class="border border-red-500 text-red-500 bg-white rounded-md p-1 hover:bg-red-500 hover:text-white" on:click={() =>
                                        data.voice[index].behavior = data.voice[index].behavior.filter((_, j) => j !== i)
                                    }>
                                        <i class="las la-trash"></i>
                                    </button>
                                    <div class="flex items-center space-x-2 w-2/5">
                                        <input type="text" class="w-full border border-gray-300 rounded p-1 mx-2" placeholder="識別子">
                                    </div>
                                    <div class="flex items-center space-x-2 w-2/5">
                                        <input type="text" class="w-full border border-gray-300 rounded p-1 mx-2" placeholder="画像パス">
                                    </div>
                                </div>
                                {/each}
                                <button class="m-2 px-2 py-1 border border-blue-500 text-blue-500 bg-white rounded-md hover:bg-blue-500 hover:text-white" on:click={() =>
                                    data.voice[index].behavior = [...data.voice[index].behavior, { identification: "", imagePath: "" }]
                                }>
                                    <i class="las la-plus"></i>
                                </button>
                            </div>
                            <p class="text-xs text-gray-500">※ imagesフォルダ内に画像を配置してください</p>
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="voice" class="text-sm">音声設定</label>
                            <select id="voice" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].type}>
                                <option value="voicevox">VOICEVOX</option>
                                <option value="bertvits2">Bert-VITS2(FastAPI)</option>
                                <option value="stylebertvits2">Style-Bert-VITS2(FastAPI)</option>
                                <option value="google-text-to-speech">Google Text to Speech API</option>
                            </select>
                        </div>
                    </div>
                    {#if data.multiVoice}
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="model_id" class="text-sm">キャラクター識別子</label>
                                <input type="text" id="model_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].identification} />
                            </div>
                        </div>
                    {/if}
                    {#if data.voice[index].type === "voicevox"}
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="speaker_id" class="text-sm">スピーカーID</label>
                                <input type="text" id="speaker_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].speakerId} />
                            </div>
                        </div>
                    {/if}
                    {#if data.voice[index].type === "bertvits2"}
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="model_id" class="text-sm">モデルID</label>
                                <input type="text" id="model_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelId} />
                            </div>
                        </div>
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="speaker_id" class="text-sm">スピーカーID</label>
                                <input type="text" id="speaker_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].speakerId} />
                            </div>
                        </div>
                    {/if}
                    {#if data.voice[index].type === "stylebertvits2"}
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="model_id" class="text-sm">モデル</label>
                            <input type="text" id="model_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelId} />
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="model_file" class="text-sm">モデルファイル</label>
                            <input type="text" id="model_file" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelFile} />
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="speaker_id" class="text-sm">話者</label>
                            <input type="text" id="speaker_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].speakerId} />
                        </div>
                    </div>
                    {/if}
                    {#if data.voice[index].type === "google-text-to-speech"}
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="gender" class="text-sm">性別</label>
                                <select id="gender" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].speakerId}>
                                    <option value="MALE">男性</option>
                                    <option value="FEMALE">女性</option>
                                    <option value="NEUTRAL">ナチュラル</option>
                                    <option value="">指定なし</option>
                                </select>
                            </div>
                        </div>
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="model_id" class="text-sm">音声名</label>
                                <input type="text" id="model_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelId} />
                            </div>
                        </div>
                    {/if}
                    {#if index > 0}
                        <div class="flex justify-between items-center p-4">
                            <button class="border border-red-500 text-red-500 bg-white rounded-md px-4 py-2 hover:bg-red-500 hover:text-white" on:click={() =>
                                data.voice = data.voice.filter((_, i) => i !== index)
                            }>
                                <i class="las la-trash"></i> 削除
                            </button>
                        </div>
                    {/if}
                </div>
            {/each}
            {#if data.multiVoice}
            <div class="flex justify-between items-center p-4">
                <button class="border border-blue-500 text-blue-500 bg-white rounded-md px-4 py-2 hover:bg-blue-500 hover:text-white" on:click={() => 
                    data.voice = [...data.voice, { type: "voicevox", modelId: "", speakerId: "1", identification: "", modelFile: "" ,image: "", backgroundImagePath: "", behavior: []}]
                }>
                    <i class="las la-plus"></i> 追加
                </button>
            </div>
            {/if}
        {/if}

        <!-- チャット設定 -->
        <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
            <i class="las la-comments text-2xl mr-2"></i>
            チャット設定
            <i class={"las text-2xl ml-auto" + (showChat ? " la-angle-up" : " la-angle-down")} on:click={() => (showChat = !showChat)}></i>
        </h2>
        {#if showChat}
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="chat" class="text-sm">チャット設定</label>
                    <select id="chat" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.type}>
                        <option value="openai">OpenAI</option>
                        <option value="anthropic">Anthropic</option>
                        <option value="cohere">Cohere</option>
                        <option value="gemini">Gemini</option>
                        <option value="openai-local">Local LLM(OpenAI v1/chat/completions Compatible)</option>
                    </select>
                </div>
            </div>
            <!-- モデル選択 -->
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="model" class="text-sm">モデル</label>
                    <input type="text" id="model" class="w-full border border-gray-300 rounded p-1" list="model_list" bind:value={data.chat.model} />
                    <datalist id="model_list">
                        {#if data.chat.type === "openai"}
                            <option value="gpt-4o">GPT-4o</option>
                            <option value="gpt-4-turbo">GPT-4 Turbo</option>
                            <option value="gpt-4-turbo-2024-04-09">GPT-4 Turbo(2024-04-09)</option>
                            <option value="gpt-4-turbo-preview">GPT-4 Turbo Preview</option>
                            <option value="gpt-4-0125-preview">GPT-4 Turbo(0125-preview)</option>
                            <option value="gpt-4-1106-preview">GPT-4 Turbo(1106-preview)</option>
                            <option value="gpt-4">GPT-4</option>
                            <option value="gpt-4-32k">GPT-4 32k</option>
                            <option value="gpt-3.5-turbo-0125">GPT-3.5 Turbo(0125)</option>
                            <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                            <option value="gpt-3.5-turbo-1106">GPT-3.5 Turbo(1106)</option>
                        {/if}
                        {#if data.chat.type === "anthropic"}
                            <option value="claude-3-opus-20240229">Claude 3 Opus(20240229)</option>
                            <option value="claude-3-sonnet-20240229">Claude 3 Sonnet(20240229)</option>
                            <option value="claude-3-haiku-20240307">Claude 3 Haiku(20240307)</option>
                        {/if}
                        {#if data.chat.type === "cohere"}
                            <option value="command-r-plus">Command R Plus</option>
                            <option value="command-r">Command R</option>
                            <option value="command-nightly">Command Nightly</option>
                            <option value="command">Command</option>
                            <option value="command-light-nightly">Command Light Nightly</option>
                            <option value="command-light">Command Light</option>
                        {/if}
                        {#if data.chat.type === "gemini"}
                            <option value="gemini-1.5-pro-latest">Gemini 1.5 Pro</option>
                            <option value="gemini-1.5-flash-latest">Gemini 1.5 Flash</option>
                            <option value="gemini-pro">Gemini Pro</option>
                        {/if}
                    </datalist>
                </div>
            </div>
            <!-- システムプロンプト -->
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="system_prompt" class="text-sm">システムプロンプト</label>
                    <textarea id="system_prompt" 
                     rows="10"
                     class="w-full border border-gray-300 rounded p-1 resize-y" bind:value={data.chat.systemPrompt}></textarea>
                </div>
            </div>
        {/if}

        <!-- チャット履歴のリセット -->
        <div class="flex justify-center items-center p-4">
            <button class="bg-red-500 text-white rounded-md p-2 w-64" on:click={onReset}>
                会話履歴をリセットする
            </button>
        </div>
        <!-- 保存/キャンセル -->
        <div class="flex justify-center items-center p-4">
            <button class={"bg-blue-500 text-white rounded-md p-2 w-24 " + (saveLoading ? " opacity-50 cursor-not-allowed" : "")} on:click={() => onSave()}>
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
