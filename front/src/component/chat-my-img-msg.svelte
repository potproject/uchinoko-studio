<script lang="ts">
    export let loading = false;
    export let message = '';
    export let image = '';
</script>

<div class="flex justify-end items-center">
    <div class="flex justify-center items-center">
        <div class="bg-blue-200 rounded-xl p-2 m-2">
            <p class="text-gray-600">
                {message}
            </p>
            <img src={image} alt="image" class="w-full h-48 object-cover rounded-xl" />
            <div class="flex justify-end items-center">
                <div class="flex justify-center items-center">
                    <i class="las la-sync-alt animate-spin text-cyan-600 text-2xl {loading ? 'opacity-100' : 'opacity-0'}"></i>
                </div>
            </div>
        </div>
    </div>
</div>