<script lang="ts">
    import { onMount, createEventDispatcher } from "svelte";
    import type { CharacterConfigList } from "../types/character";

    const dispatch = createEventDispatcher();

    export let characters: CharacterConfigList;

    export let selectCharacterIndex: number | undefined;
</script>

<!-- キャラクター画像と名前 -->
<div class="overflow-x-auto flex items-center bg-gray-100 py-4 px-2 w-full mt-2">
    <!-- 10個のキャラクターを表示する -->
    {#each characters.characters as character, i}
        <div class="flex items-center px-2 mx-6 flex-col w-28">
            <!-- キャラクター画像をクリックすると、キャラクター設定モーダルを表示する -->
            <div class="relative">
                <img
                    src={character.general.image}
                    alt={character.general.name}
                    class={"w-24 h-24 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2 " + (selectCharacterIndex === i ? "border-blue-500 shadow-md" : "border-gray-300")}
                    on:click={() => (dispatch("selectCharacter", { index: i }))}
                />
                <i
                    on:click={() => window.confirm("このキャラクターを削除しますか？") && dispatch("deleteCharacter", { character: character })}
                    class="las la-trash-alt text-gray-200 absolute -left-4 -top-1 text-3xl bg-red-500 rounded-full cursor-pointer p-1 hover:bg-red-700 hover:text-white"
                ></i>
                <i
                    on:click={() => (dispatch("openCharacterConfig", { character: character }))}
                    class="las la-wrench text-gray-500 absolute -right-4 -top-1 text-3xl bg-white rounded-full cursor-pointer border border-gray-300 p-1 hover:bg-gray-100 hover:text-blue-500"
                ></i>
            </div>

            <!-- キャラクター名 -->
            <div class={"text-sm truncate w-28 mt-3 text-center " + (selectCharacterIndex === i ? "text-blue-500" : "text-gray-500")}>{character.general.name}</div>
        </div>
    {/each}
    <!-- Add Btn -->
    <div class="flex items-center px-2 mx-2 flex-col w-28">
        <div class="bg-white w-12 h-12 mx-12 my-10 rounded-full border shadow-sm bg-white cursor-pointer hover:shadow-md border-2 border-gray-300 flex items-center justify-center"
            on:click={() => dispatch("openCharacterConfig", { character: null })}
        >
            <i class="las la-plus text-4xl text-gray-300 cursor-pointer hover:text-blue-500"></i>
        </div>
    </div> 
</div>
