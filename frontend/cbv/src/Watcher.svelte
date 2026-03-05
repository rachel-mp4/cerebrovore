<script lang="ts">
  import { WatcherContext } from "./lib/watchercontext.svelte";
  import WatchedThreads from "./lib/components/WatchedThreads.svelte";
  const wc = new WatcherContext();
</script>

<WatchedThreads items={wc.watchthreads} />
