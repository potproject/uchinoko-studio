<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import type { CharacterConfig, MemoryItem, MemoryItemList, SessionSummary } from "../../types/character";
    import type { ChatSessionList } from "../../types/chat";
    import { getID } from "$lib/GetId";
    const dispatch = createEventDispatcher();

    let showVoice = false;
    let showChat = false;
    let showMemory = false;

    let saveLoading = false;
    let memoryLoading = false;
    let memoryItems: MemoryItem[] = [];
    let sessionSummary: SessionSummary | null = null;
    const ownerId = getID();

    export let data: CharacterConfig;

    const loadMemory = async () => {
        memoryLoading = true;
        try {
            const itemsRes = await fetch(`/v1/memory/${ownerId}/${data.general.id}/items`);
            if (itemsRes.ok) {
                const items = await itemsRes.json() as MemoryItemList;
                memoryItems = items.items;
            }
            const summaryRes = await fetch(`/v1/memory/${ownerId}/${data.general.id}/session/${ownerId}/summary`);
            if (summaryRes.ok) {
                sessionSummary = await summaryRes.json() as SessionSummary;
            }
        } catch (e) {
            console.error(e);
        } finally {
            memoryLoading = false;
        }
    };

    onMount(() => {
        loadMemory();
    });

    const onReset = async () => {
        if (window.confirm("チャット履歴をリセットしますか？")) {
            const ownerId = getID();
            const list = await fetch(`/v1/chat/${ownerId}/${data.general.id}/sessions`, {
                method: "GET",
            }).then((res) => res.json() as Promise<ChatSessionList>);

            await Promise.all(list.sessions.map((session) => {
                const query = session.sessionId === ownerId ? "" : `?${new URLSearchParams({ sessionId: session.sessionId }).toString()}`;
                return fetch(`/v1/chat/${ownerId}/${data.general.id}${query}`, {
                    method: "DELETE",
                });
            }));

            alert("チャット履歴をリセットしました");
            location.reload();
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

    const createMemoryItem = async (scope: "character" | "relationship") => {
        const res = await fetch(`/v1/memory/${ownerId}/${data.general.id}/items`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                scope,
                kind: scope === "character" ? "persona_rule" : "relationship_fact",
                content: "",
                keywordsText: "",
                pinned: scope === "character",
                confidence: 1,
                salience: 1,
            }),
        });
        if (!res.ok) {
            window.alert("Memory の追加に失敗しました");
            return;
        }
        const created = await res.json() as MemoryItem;
        memoryItems = [created, ...memoryItems];
    };

    const saveMemoryItem = async (item: MemoryItem) => {
        const res = await fetch(`/v1/memory/item/${item.id}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(item),
        });
        if (!res.ok) {
            window.alert("Memory の保存に失敗しました");
            return;
        }
        const updated = await res.json() as MemoryItem;
        memoryItems = memoryItems.map((v) => v.id === updated.id ? updated : v);
    };

    const deleteMemoryItem = async (item: MemoryItem) => {
        const res = await fetch(`/v1/memory/item/${item.id}`, {
            method: "DELETE",
        });
        if (!res.ok) {
            window.alert("Memory の削除に失敗しました");
            return;
        }
        memoryItems = memoryItems.filter((v) => v.id !== item.id);
    };

    const rebuildMemory = async () => {
        const res = await fetch(`/v1/memory/${ownerId}/${data.general.id}/rebuild`, {
            method: "POST",
        });
        if (!res.ok) {
            window.alert("Memory の再構築に失敗しました");
            return;
        }
        window.alert("Memory の再構築ジョブを登録しました");
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
                <label for="name" class="text-sm">名前</label>
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
                        <div class="flex-1">
                            <label for="name" class="text-sm">キャラクター名</label>
                            <input type="text" id="name" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].name} />
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <!-- キャラクター画像 -->
                        <div class="flex-1">
                            <label for="image" class="text-sm">キャラクターアイコン画像</label>
                            <img src={"images/"+data.voice[index].image} alt="キャラクター画像" class="w-24 h-24 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2" />
                            <div class="flex items-center py-1">
                                <p class="text-xs text-gray-500">images/</p>
                                <input type="text" id="image" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].image} placeholder="画像パス" />
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="backgroundImage" class="text-sm">立ち絵画像</label>
                            <div class="flex items-center py-1">
                                <p class="text-xs text-gray-500">images/</p>
                                <input type="text" id="backgroundImage" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].backgroundImagePath} placeholder="画像パス" />
                            </div>
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
                                        <p class="text-xs text-gray-500">images/</p>
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
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="voice" class="text-sm">音声設定</label>
                            <select id="voice" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].type}>
                                <option value="voicevox">VOICEVOX</option>
                                <option value="bertvits2">Bert-VITS2(FastAPI)</option>
                                <option value="irodori-tts">Irodori TTS(Gradio)</option>
                                <option value="stylebertvits2">Style-Bert-VITS2(FastAPI)</option>
                                <option value="nijivoice">にじボイス API</option>
                                <option value="google-text-to-speech">Google Text to Speech API</option>
                                <option value="openai-speech">OpenAI Speech API</option>
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
                    {#if data.voice[index].type === "irodori-tts"}
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="model_id" class="text-sm">Checkpoint</label>
                                <input type="text" id="model_id" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelId} placeholder="Aratako/Irodori-TTS-500M-v2" />
                            </div>
                        </div>
                        <div class="flex items-center px-4 py-2">
                            <div class="flex-1">
                                <label for="reference_audio_path" class="text-sm">参照音声URLまたは refs/ 配下パス</label>
                                <input type="text" id="reference_audio_path" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].referenceAudioPath} placeholder="sample.wav または https://example.com/sample.wav" />
                                <p class="text-xs text-gray-500 mt-1">空欄なら参照音声なし。ローカル音声はリポジトリ直下の refs/ に置いてファイル名を指定します。</p>
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
                    {#if data.voice[index].type === "nijivoice"}
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="speaker_id" class="text-sm">ボイスアクターID</label>
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
                    {#if data.voice[index].type === "openai-speech"}
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="model" class="text-sm">モデル</label>
                            <select id="model" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].modelId}>
                                <option value="tts-1">tts-1</option>
                                <option value="tts-1-hd">tts-1-hd</option>
                            </select>
                        </div>
                    </div>
                    <div class="flex items-center px-4 py-2">
                        <div class="flex-1">
                            <label for="voice" class="text-sm">ボイス</label>
                            <select id="voice" class="w-full border border-gray-300 rounded p-1" bind:value={data.voice[index].speakerId}>
                                <option value="alloy">alloy</option>
                                <option value="echo">echo</option>
                                <option value="fable">fable</option>
                                <option value="onyx">onyx</option>
                                <option value="nova">nova</option>
                                <option value="shimmer">shimmer</option>
                            </select>
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
                    data.voice = [...data.voice, { name:"", type: "voicevox", modelId: "", speakerId: "1", identification: "", modelFile: "", referenceAudioPath: "", image: "", backgroundImagePath: "", behavior: []}]
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
                        <option value="deepseek">DeepSeek</option>
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
                            <option value="gpt-5.2">GPT-5.2</option>
                            <option value="gpt-5-mini">GPT-5 Mini</option>
                            <option value="gpt-5-nano">GPT-5 Nano</option>
                            <option value="gpt-5">GPT-5</option>
                        {/if}
                        {#if data.chat.type === "anthropic"}
                            <option value="claude-opus-4-5">Claude 4.5 Opus</option>
                            <option value="claude-sonnet-4-5">Claude 4.5 Sonnet</option>
                            <option value="claude-haiku-4-5">Claude 4.5 Haiku</option>
                        {/if}
                        {#if data.chat.type === "deepseek"}
                            <option value="deepseek-chat">DeepSeek Chat</option>
                            <option value="deepseek-reasoner">DeepSeek Reasoner</option>
                        {/if}
                        {#if data.chat.type === "gemini"}
                            <option value="gemini-3-pro-preview">Gemini 3 Pro Preview</option>
                            <option value="gemini-3-flash-preview">Gemini 3 Flash Preview</option>
                            <option value="gemini-2.5-flash">Gemini 2.5 Flash</option>
                            <option value="gemini-2.5-flash-lite">Gemini 2.5 Flash Lite</option>
                            <option value="gemini-2.5-pro">Gemini 2.5 Pro</option>
                        {/if}
                    </datalist>
                </div>
            </div>
            <!-- Temperature -->
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="temperature" class="text-sm">Temperature</label>
                    <div class="flex items center">
                        <input type="checkbox" id="temperature" class="mr-2" bind:checked={data.chat.temperature.enable} />
                        <input type="number" min="0" max="2" step="0.01" id="temperature" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.temperature.value} disabled={!data.chat.temperature.enable} />
                    </div>
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
            <!-- 最大履歴保持数 -->
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="max_history" class="text-sm">最大履歴保持数(0で無限)</label>
                    <input type="number" id="max_history" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.maxHistory} />
                </div>
            </div>
            <!-- レートリミット -->
            <div class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">レートリミット(0で無効)</div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">リクエスト(1日)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.day.request} />
                </div>
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">トークン(1日)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.day.token} />
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">リクエスト(1時間)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.hour.request} />
                </div>
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">トークン(1時間)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.hour.token} />
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">リクエスト(1分)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.minute.request} />
                </div>
                <div class="flex-1">
                    <label for="rate_limit" class="text-sm">トークン(1分)</label>
                    <input type="number" id="rate_limit" class="w-full border border-gray-300 rounded p-1" bind:value={data.chat.limit.minute.token} />
                </div>
            </div>
        {/if}

        <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
            <i class="las la-brain text-2xl mr-2"></i>
            Memory
            <i class={"las text-2xl ml-auto" + (showMemory ? " la-angle-up" : " la-angle-down")} on:click={() => (showMemory = !showMemory)}></i>
        </h2>
        {#if showMemory}
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <input type="checkbox" id="memory_enabled" class="mr-2" bind:checked={data.memory.enabled} />
                    <label for="memory_enabled" class="text-sm">Memory を有効にする</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="memory_max_items" class="text-sm">Prompt に入れる最大件数</label>
                    <input type="number" id="memory_max_items" class="w-full border border-gray-300 rounded p-1" bind:value={data.memory.maxItemsInPrompt} />
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <input type="checkbox" id="memory_relationship" class="mr-2" bind:checked={data.memory.enableRelationshipMemory} />
                    <label for="memory_relationship" class="text-sm">Relationship Memory を有効にする</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <input type="checkbox" id="memory_summary" class="mr-2" bind:checked={data.memory.enableSessionSummary} />
                    <label for="memory_summary" class="text-sm">Session Summary を有効にする</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <input type="checkbox" id="memory_semantic" class="mr-2" bind:checked={data.memory.enableSemanticSearch} />
                    <label for="memory_semantic" class="text-sm">Semantic Search を有効にする</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <input type="checkbox" id="memory_sensitive" class="mr-2" bind:checked={data.memory.allowSensitiveMemory} />
                    <label for="memory_sensitive" class="text-sm">センシティブ情報の保存を許可する</label>
                </div>
            </div>
            <div class="flex items-center px-4 py-2">
                <div class="flex-1">
                    <label for="memory_embedding" class="text-sm">Embedding Model</label>
                    <input type="text" id="memory_embedding" class="w-full border border-gray-300 rounded p-1" bind:value={data.memory.embeddingModel} />
                </div>
            </div>

            <div class="px-4 py-2">
                <div class="flex items-center mb-2">
                    <div class="text-sm font-bold flex-1">Session Summary</div>
                    <button class="border border-blue-500 text-blue-500 bg-white rounded-md px-3 py-1 hover:bg-blue-500 hover:text-white" on:click={rebuildMemory}>Rebuild</button>
                </div>
                <textarea class="w-full border border-gray-300 rounded p-2 resize-y" rows="5" readonly>{sessionSummary?.summary ?? ""}</textarea>
            </div>

            <div class="px-4 py-2">
                <div class="flex items-center mb-2">
                    <div class="text-sm font-bold flex-1">Manual Memory</div>
                    <button class="border border-blue-500 text-blue-500 bg-white rounded-md px-3 py-1 mr-2 hover:bg-blue-500 hover:text-white" on:click={() => createMemoryItem("character")}>Character</button>
                    <button class="border border-green-500 text-green-500 bg-white rounded-md px-3 py-1 hover:bg-green-500 hover:text-white" on:click={() => createMemoryItem("relationship")}>Relationship</button>
                </div>
                {#if memoryLoading}
                    <div class="text-sm text-gray-500">Memory を読み込み中です...</div>
                {/if}
                {#each memoryItems as item}
                    <div class="border border-gray-300 rounded p-3 mb-3">
                        <div class="flex items-center mb-2">
                            <select class="border border-gray-300 rounded p-1 mr-2" bind:value={item.scope} disabled>
                                <option value="character">character</option>
                                <option value="relationship">relationship</option>
                            </select>
                            <input type="text" class="flex-1 border border-gray-300 rounded p-1 mr-2" bind:value={item.kind} placeholder="kind" />
                            <label class="text-xs mr-2"><input type="checkbox" class="mr-1" bind:checked={item.pinned} />pin</label>
                            <button class="border border-blue-500 text-blue-500 bg-white rounded-md px-3 py-1 mr-2 hover:bg-blue-500 hover:text-white" on:click={() => saveMemoryItem(item)}>保存</button>
                            <button class="border border-red-500 text-red-500 bg-white rounded-md px-3 py-1 hover:bg-red-500 hover:text-white" on:click={() => deleteMemoryItem(item)}>削除</button>
                        </div>
                        <textarea class="w-full border border-gray-300 rounded p-2 resize-y mb-2" rows="4" bind:value={item.content} placeholder="memory content"></textarea>
                        <input type="text" class="w-full border border-gray-300 rounded p-1 mb-2" bind:value={item.keywordsText} placeholder="keywords" />
                        <div class="flex items-center">
                            <div class="flex-1 mr-2">
                                <label class="text-xs">confidence</label>
                                <input type="number" min="0" max="1" step="0.01" class="w-full border border-gray-300 rounded p-1" bind:value={item.confidence} />
                            </div>
                            <div class="flex-1">
                                <label class="text-xs">salience</label>
                                <input type="number" min="0" max="1" step="0.01" class="w-full border border-gray-300 rounded p-1" bind:value={item.salience} />
                            </div>
                        </div>
                    </div>
                {/each}
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
