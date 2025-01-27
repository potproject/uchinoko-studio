<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { onMount } from 'svelte';
    const dispatch = createEventDispatcher();

    type EnvVarItem = {
        key: string;
        name: string;
        description: string;
        link?: string;
        value: string;
        isSet: boolean;
    };

    type Category = {
        name: string;
        icon: string;
        items: EnvVarItem[];
    };

    let categories: Category[] = [
        {
            name: "STT API (Speech to Text)",
            icon: "las la-microphone",
            items: [
                {
                    key: "OPENAI_SPEECH_TO_TEXT_API_KEY",
                    name: "OpenAI Speech to Text API Key",
                    description: "OpenAI Speech to text(音声認識)サービス用APIキー",
                    link: "https://platform.openai.com/docs/guides/speech-to-text",
                    value: "",
                    isSet: false,
                },
                {
                    key: "GOOGLE_SPEECH_TO_TEXT_API_KEY",
                    name: "Google Speech-to-Text API Key",
                    description: "Google Speech-to-Text(音声認識)サービス用APIキー",
                    link: "https://cloud.google.com/speech-to-text/docs",
                    value: "",
                    isSet: false,
                },
                {
                    key: "VOSK_SERVER_ENDPOINT",
                    name: "Voskサーバーエンドポイント",
                    description: "Vosk音声認識サーバーのエンドポイントURL",
                    link: "https://alphacephei.com/vosk/server",
                    value: "",
                    isSet: false,
                },
            ],
        },
        {
            name: "Chat API",
            icon: "las la-comments",
            items: [
                {
                    key: "OPENAI_API_KEY",
                    name: "OpenAI API Key",
                    description: "OpenAIのChat用APIキー",
                    link: "https://platform.openai.com/docs/guides/chat-completions",
                    value: "",
                    isSet: false,
                },
                {
                    key: "ANTHROPIC_API_KEY",
                    name: "Anthropic API Key",
                    description: "AnthropicのChat用APIキー",
                    link: "https://docs.anthropic.com/en/home",
                    value: "",
                    isSet: false,
                },
                {
                    key: "DEEPSEEK_API_KEY",
                    name: "DeepSeek API Key",
                    description: "DeepSeekのChat用APIキー",
                    link: "https://api-docs.deepseek.com/",
                    value: "",
                    isSet: false,
                },
                {
                    key: "GEMINI_API_KEY",
                    name: "Gemini API Key",
                    description: "GeminiのChat用APIキー",
                    link: "https://ai.google.dev/gemini-api/docs/api-key",
                    value: "",
                    isSet: false,
                },
            ],
        },
        {
            name: "Chat Local API/Endpoint",
            icon: "las la-server",
            items: [
                {
                    key: "OPENAI_LOCAL_API_KEY",
                    name: "OpenAIローカルAPIキー",
                    description: "ローカルOpenAI API用のAPIキー",
                    link: "",
                    value: "",
                    isSet: false,
                },
                {
                    key: "OPENAI_LOCAL_API_ENDPOINT",
                    name: "OpenAIローカルAPIエンドポイント",
                    description: "ローカルOpenAI APIのエンドポイントURL(Ollamaなど)",
                    link: "",
                    value: "",
                    isSet: false,
                },
            ],
        },
        {
            name: "TTS API/Endpoint (Text to Speech)",
            icon: "las la-volume-up",
            items: [
                {
                    key: "VOICEVOX_ENDPOINT",
                    name: "VOICEVOXエンドポイント",
                    description: "音声合成ソフトウェア VOICEVOX のエンドポイントURL",
                    link: "https://voicevox.hiroshiba.jp",
                    value: "",
                    isSet: false,
                },
                {
                    key: "BERTVITS2_ENDPOINT",
                    name: "BERTVITS2エンドポイント",
                    description: "音声合成OSS Bert-VITS2 のエンドポイントURL",
                    link: "https://github.com/fishaudio/Bert-VITS2",
                    value: "",
                    isSet: false,
                },
                {
                    key: "STYLEBERTVIT2_ENDPOINT",
                    name: "STYLEBERTVIT2エンドポイント",
                    description: "音声合成OSS Style-Bert-VITS2 のエンドポイントURL",
                    link: "https://github.com/litagin02/Style-Bert-VITS2",
                    value: "",
                    isSet: false,
                },
                {
                    key: "NIJIVOICE_API_KEY",
                    name: "にじボイスAPIキー",
                    description: "Style-Bert-VITS2の音声合成サービス用APIキー",
                    link: "https://nijivoice.com",
                    value: "",
                    isSet: false,
                },
                {
                    key: "GOOGLE_TEXT_TO_SPEECH_API_KEY",
                    name: "Google Text to Speech API Key",
                    description: "GoogleのText-to-Speech(音声合成)サービス用APIキー",
                    link: "https://cloud.google.com/text-to-speech/docs",
                    value: "",
                    isSet: false,
                },
                {
                    key: "OPENAI_SPEECH_API_KEY",
                    name: "OpenAI Speech API Key",
                    description: "OpenAIのSpeech to text(音声合成)サービス用APIキー",
                    link: "https://platform.openai.com/docs/guides/speech-to-text",
                    value: "",
                    isSet: false,
                },
            ],
        },
    ];

    let saveLoading = false;

    // サーバーから既存の環境変数の状態を取得
    onMount(() => {
        fetch('/v1/config/env')
            .then(res => res.json())
            .then(envData => {
                const cloneCategories = structuredClone(categories);
                cloneCategories.forEach(category => {
                    category.items.forEach(item => {
                        if (envData[item.key]) {
                            item.isSet = true;
                        }
                    });
                });
                categories = cloneCategories;
            })
            .catch(err => {
                console.error('環境変数の状態取得に失敗しました', err);
            });
    });

    let onSave = () => {
        saveLoading = true;
        // 入力された値を収集
        let envVars: { [key: string]: string } = {};
        categories.forEach(category => {
            category.items.forEach(item => {
                if (item.value) {
                    envVars[item.key] = item.value;
                }
            });
        });
        fetch(`/v1/config/env`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(envVars),
        })
            .then((res) => {
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
                    <i class="las la-database text-2xl mr-2"></i>
                    環境変数設定
                </h1>
                <div class="flex items-center text-gray-300 hover:text-gray-800 cursor-pointer" on:click={() => dispatch("close")}>
                    <i class="las la-times text-2xl"></i>
                </div>
            </div>
            {#each categories as category}
                <h2 class="text-xl px-2 py-2 border-b border-gray-300 flex items-center mb-2 mx-4 mt-4">
                    <i class="{category.icon} text-2xl mr-2"></i>
                    {category.name}
                </h2>
                {#each category.items as item}
                    <div class={"px-4 py-2 mx-4 my-2 border border-gray-200 rounded " + (item.isSet ? "bg-green-50" : "")}>
                        <div class="flex justify-between items-center">
                            <label class="text-sm font-semibold">{item.name}</label>
                            <div class="flex items-center">
                                {#if item.link}
                                    <a href="{item.link}" target="_blank" class="text-blue-500 hover:text-blue-700 mr-2">
                                        <i class="las la-external-link-alt text-xl"></i>
                                    </a>
                                {/if}
                                {#if item.isSet}
                                    <i class="las la-check text-green-500 text-xl"></i>
                                {/if}
                            </div>
                        </div>
                        <p class="text-xs text-gray-500 mb-2">{item.description}</p>
                        <input type="password" class="w-full border border-gray-300 rounded p-1" bind:value={item.value} placeholder="値を入力">
                    </div>
                {/each}
            {/each}
            <!-- 保存/キャンセル -->
            <div class="flex justify-center items-center p-4">
                <button class={"bg-blue-500 text-white rounded-md p-2 w-24 " + (saveLoading ? " opacity-50 cursor-not-allowed" : "")} on:click={() => onSave()}>
                    {#if saveLoading}
                        <i class="las la-spinner animate-spin"></i>
                    {:else}
                        更新
                    {/if}
                </button>
                <button class="bg-gray-300 text-gray-800 rounded-md p-2 w-24 ml-2" on:click={() => dispatch("close")}>キャンセル</button>
            </div>
        </div>
    </div>
</div>
