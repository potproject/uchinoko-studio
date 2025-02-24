<script lang="ts">
    export let message = '';
    export let image;
</script>

<div class="flex justify-end items-center">
    <div class="flex justify-center items-center">
        <div class="bg-blue-200 rounded-xl p-2 m-2">
            <p class="text-gray-600">{message}</p>
            {#if image}
            <img src={image} alt="image" class="w-full h-48 object-cover rounded-xl" />
            {/if}
        </div>
    </div>
</div>