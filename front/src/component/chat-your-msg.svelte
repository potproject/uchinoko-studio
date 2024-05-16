<script lang="ts">
    export let loading = false;
    export let speaking = false;
    export let message = '';
    export let img: string|null = 'default.png';
</script>

<div class="flex justify-start items-center">
    <div class="flex justify-center items-start">
        {#if img}
        <img src={img} class="rounded-full w-12 h-12 mt-4 ml-2" alt="avatar" />
        {/if}
        <div class="bg-gray-200 rounded-xl p-2 m-2">
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