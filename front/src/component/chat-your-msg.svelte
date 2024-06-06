<script lang="ts">
    export let loading = false;
    export let speaking = false;
    export let name = '';
    export let message = '';
    export let img: string|null = 'default.png';
</script>

<div class="flex justify-start items-center">
    <div class="flex justify-center items-start">
        {#if img}
        <img src={img} class="rounded-full w-12 h-12 mt-2 ml-2" alt="avatar" />
        {/if}
        <div class="flex flex-col">
            <div class="flex items-center">
                <p class="text-gray-500 font-bold text-sm pl-2 pt-2">{name}</p>
            </div>
            <div class="bg-gray-200 rounded-xl p-2 mx-2 mt-1 mb-2">
                <p class="text-gray-600 whitespace-pre-wrap">{message}</p>
                <div class="flex justify-end items-center">
                    <div class="flex justify-center items-center">
                        <i class="las la-sync-alt animate-spin text-cyan-600 text-2xl {loading ? 'opacity-100' : 'opacity-0'}"></i>
                        <i class="las la-phone-volume text-cyan-600 text-2xl {speaking ? 'opacity-100' : 'opacity-0'}"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>