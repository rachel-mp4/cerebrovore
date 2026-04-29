<script lang="ts">
  import type { WatchThread } from "../types";
  import WatchedThread from "./WatchedThread.svelte";
  import NewThread from "./NewThread.svelte";
  interface Props {
    isforum: boolean;
    items: Array<WatchThread>;
    newitems: Array<WatchThread>;
    watchIdx: (idx: number) => void;
    rmIdx: (idx: number) => void;
  }
  let { isforum, items, newitems, watchIdx, rmIdx }: Props = $props();
  let bumpOrder = $derived(
    items
      .toSorted((a: WatchThread, b: WatchThread) => b.bumpedAt - a.bumpedAt)
      .slice(0, 10),
  );
  let bumpsOrder = $derived(
    items
      .toSorted((a: WatchThread, b: WatchThread) => b.bumps - a.bumps)
      .slice(0, 10),
  );
  let usingBumpOrdering = $state(false);
  const button = document.querySelector(
    "#watched-threads button",
  ) as HTMLButtonElement;
  button.onclick = () => {
    button.classList.toggle("recent-activity");
    usingBumpOrdering = button.classList.toggle("total-activity");
  };
  let order = $derived(usingBumpOrdering ? bumpsOrder : bumpOrder);
</script>

{#if items.length === 0}
  <div>nothing so far...</div>
{:else}
  {#each order as item}
    <WatchedThread {item} />
  {/each}
{/if}
{#if newitems.length !== 0}
  <div>new threads</div>
  {#each newitems.slice(0, 3) as item, idx}
    <NewThread {item} {idx} {watchIdx} {rmIdx} />
  {/each}
{/if}
