<script lang="ts">
  import { WormWatchContext } from "./lib/wormwatchcontext.svelte";
  import WormWatch from "./lib/components/WormWatch.svelte";
  const ctx = new WormWatchContext();
  const nww = localStorage.getItem("nowormwatch");
</script>

{#if nww === null || nww === ""}
  <WormWatch {ctx} />
{/if}
