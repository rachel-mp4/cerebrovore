<script lang="ts">
  import { WatcherContext } from "./lib/watchercontext.svelte";
  import WatchedThreads from "./lib/components/WatchedThreads.svelte";
  interface Props {
    isforum: boolean;
  }
  let { isforum }: Props = $props();
  const wc = new WatcherContext(isforum);
</script>

<WatchedThreads
  {isforum}
  items={wc.watchthreads}
  newitems={wc.newthreads}
  watchIdx={wc.watchIdx.bind(wc)}
  rmIdx={wc.rmIdx.bind(wc)}
/>
