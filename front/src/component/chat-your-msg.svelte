<script lang="ts">
    export let loading = false;
    export let speaking = false;
    export let message = '';
    export let img = 'default.png';
</script>

<div>
    <div class="flex justify-center items-start">
        <img src={img} class="rounded-full w-6 h-6 mt-4 ml-2" alt="avatar" />
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