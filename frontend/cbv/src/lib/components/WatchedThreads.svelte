<script lang="ts">
  import type { WatchThread } from "../types";
  import WatchedThread from "./WatchedThread.svelte";
  interface Props {
    items: Array<WatchThread>;
  }
  let { items }: Props = $props();
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
