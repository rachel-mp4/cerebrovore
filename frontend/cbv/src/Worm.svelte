<script lang="ts">
  import { WormWatchContext } from "./lib/wormwatchcontext.svelte";
  import WormWatch from "./lib/components/WormWatch.svelte";
  const url = window.location.href;
  // i think this should work for both http->ws and https->wss schemes, that's
  // why the magic number 4 is there
  const sansproto = url.slice(4);
  // i split on hashtag so that way user can load a page with a hashtag in it
  // (meaning jump to id)
  // and have the /ws not be interpreted as part of the hashtag
  const address = `ws${sansproto.split("#")[0]}/ww`;
  const ctx = new WormWatchContext(address);
  const nww = localStorage.getItem("nowormwatch");
</script>

{#if nww === null || nww === ""}
  <WormWatch {ctx} />
{/if}
