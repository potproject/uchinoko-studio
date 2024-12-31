<script lang="ts">
    import { onMount, createEventDispatcher } from "svelte";
    import type { CharacterConfig, CharacterConfigList } from "../types/character";

    const dispatch = createEventDispatcher();

    let showInfo: CharacterConfig | undefined = undefined;

    export let characters: CharacterConfigList;

    export let selectCharacterIndex: number | undefined;
</script>

<!-- キャラクター画像と名前 -->
<div class="overflow-x-auto flex items-center bg-gray-100 py-4 px-2 w-full mt-2">
    {#each characters.characters as character, i}
        <div class="flex items-center px-2 mx-4 flex-col w-28">
            <!-- キャラクター画像をクリックすると、キャラクター設定モーダルを表示する -->
            <div class="relative">
                <img
                    src={character.general.name === "3人と話す" ? "images/three.png" : "images/"+character.voice[0].image }
                    alt={character.general.name}
                    class={"w-24 h-24 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2 " + (selectCharacterIndex === i ? "border-blue-500 shadow-md" : "border-gray-300")}
                    on:click={() => (dispatch("selectCharacter", { index: i }))}
                />
                <!-- <i
                    on:click={() => window.confirm("このキャラクターを削除しますか？") && dispatch("deleteCharacter", { character: character })}
                    class="las la-trash-alt text-gray-200 absolute -left-4 -top-1 text-3xl bg-red-500 rounded-full cursor-pointer p-1 hover:bg-red-700 hover:text-white"
                ></i> -->
                <!-- <i
                    on:click={() => (dispatch("openCharacterConfig", { character: character }))}
                    class="las la-wrench text-gray-500 absolute -right-4 -top-1 text-3xl bg-white rounded-full cursor-pointer border border-gray-300 p-1 hover:bg-gray-100 hover:text-blue-500"
                ></i> -->
                <i
                    on:click={() => (showInfo = character)}
                    class="las la-info-circle text-gray-500 absolute -right-4 -top-1 text-3xl bg-white rounded-full cursor-pointer border border-gray-300 p-1 hover:bg-gray-100 hover:text-blue-500"
                ></i>
            </div>

            <div class={"text-sm truncate w-28 mt-3 text-center " + (selectCharacterIndex === i ? "text-blue-500" : "text-gray-500")}>{character.general.name}</div>
        </div>
        {#if showInfo === character}
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-gray-800 bg-opacity-50">
            <div class="bg-white w-128 p-4 rounded-lg shadow-lg">
                <div class="flex items-center justify-between mb-4">
                    <h2 class="text-xl font-bold">{character.general.name}</h2>
                    <i class="las la-times text-gray-500 cursor-pointer text-xl" on:click={() => (showInfo = undefined)} ></i>
                </div>
                <hr class="border-gray-300 mb-4" />
                <div>
                    {#if character.general.name === "デフォルト" }
                    <div class="w-full flex items-center">
                        <img src={"images/"+character.voice[0].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: デフォルト(コードネーム:ai-default-01)<br>
                            性格: 前向き、元気、純粋、子供っぽい<br>
                            好きな物: おいしいもの、ゲーム、アニメ、漫画、おしゃべり<br>
                            カスタムとの関係は、双子の姉。
                            あなたのことをマスターと呼ぶ。
                        </div>
                    </div>
                    {/if}
                    {#if character.general.name === "カスタム"}
                    <div class="w-full flex items-center">
                        <img src={"images/"+character.voice[0].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: カスタム(コードネーム:ai-custom-02)<br>
                            性格: 物静か、大人しい、皮肉屋、クール<br>
                            好きな物: 静かな場所、読書、映画<br>
                            デフォルトとの関係は、双子の妹。
                            あなたのことをご主人と呼ぶ。
                        </div>
                    </div>
                    {/if}
                    {#if character.general.name === "フォーク"}
                    <div class="w-full flex items-center">
                        <img src={"images/"+character.voice[0].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: フォーク(コードネーム:ai-fork-03)<br>
                            性格: 怠惰、マイペース、睡眠欲が強い<br>
                            好きな物: 寝ること、なんとなくテレビを見ること、寝ること<br>
                            2人との関係は、たぶん末の妹。
                            あなたのことをマスターと呼ぶかもしれない。
                        </div>
                    </div>
                    {/if}
                    {#if character.general.name === "3人と話す"}
                    <div class="w-full flex items-center">
                        3人のアシスタントAIと同時に話すことができます。
                    </div>
                    <div class="w-full flex items-center">
                        <img src={"images/"+character.voice[0].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: デフォルト(コードネーム:ai-default-01)<br>
                            性格: 前向き、元気、純粋、子供っぽい<br>
                            双子の姉。
                        </div>
                    </div>  
                    <div class="w-full flex items-center mt-4">
                        <img src={"images/"+character.voice[1].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: カスタム(コードネーム:ai-custom-02)<br>
                            性格: 物静か、大人しい、皮肉屋、クール<br>
                            双子の妹。
                        </div>
                    </div>
                    <div class="w-full flex items-center mt-4">
                        <img src={"images/"+character.voice[2].backgroundImagePath} alt={character.general.name} class="w-1/4 h-full rounded-lg" />
                        <div class="w-3/4 ml-4">
                            アシスタントAI: フォーク(コードネーム:ai-fork-03)<br>
                            性格: 怠惰、マイペース、睡眠欲が強い<br>
                            たぶん妹。
                        </div>
                    </div>  
                    {/if}  
                </div>
            </div>
        </div>
        {/if}

    {/each}
    <!-- Add Btn -->
    <!--
    <div class="flex items-center px-2 mx-2 flex-col w-28">
        <div class="bg-white w-12 h-12 mx-12 my-10 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2 border-gray-300 flex items-center justify-center"
            on:click={() => dispatch("openCharacterConfig", { character: null })}
        >
            <i class="las la-plus text-4xl text-gray-300 cursor-pointer hover:text-blue-500"></i>
        </div>
    </div>  -->
</div>
