<script>
    export let text;
  </script>
  
  <div class="relative inline-block">
    <div class="tooltip-trigger">
      <slot></slot>
    </div>
    <div class="tooltip">
      {text}
    </div>
  </div>
  
  <style lang="postcss">
    .tooltip {
      @apply invisible absolute bottom-full left-1/2 mb-2 -translate-x-1/2 whitespace-nowrap rounded bg-gray-800 px-2 py-1 text-sm text-white opacity-0 transition-opacity duration-300;
    }
  
    .tooltip::after {
      content: "";
      @apply absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-800;
    }
  
    .tooltip-trigger:hover + .tooltip {
      @apply visible opacity-100;
    }
  </style>