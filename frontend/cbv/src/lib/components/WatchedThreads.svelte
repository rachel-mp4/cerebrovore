<script lang="ts">
  import { b36encodenumber } from "../utils";
  import type { WatchThread } from "../types";
  interface Props {
    items: Array<WatchThread>;
  }
  let { items }: Props = $props();
  let bumpOrder = $derived(
    items.toSorted((a: WatchThread, b: WatchThread) => b.bumpedAt - a.bumpedAt),
  );
  let bumpsOrder = $derived(
    items.toSorted((a: WatchThread, b: WatchThread) => b.bumps - a.bumps),
  );
  let usingBumpOrdering = $state(true);
  const button = document.querySelector(
    "#watched-threads button",
  ) as HTMLButtonElement;
  button.onclick = () => {
    usingBumpOrdering = button.classList.toggle("recent");
  };
</script>

{#if items.length === 0}
  <div>nothing so far...</div>
{:else if usingBumpOrdering}
  {#each bumpOrder as item}
    <div class="watched-thread">
      <div class="bump-count">{item.bumps}</div>
      <div class="thread-link">
        <a href="/t/{b36encodenumber(item.id)}">
          {item.topic ?? `#${b36encodenumber(item.id)}`}
        </a>
      </div>
    </div>
  {/each}
{:else}
  {#each bumpsOrder as item}
    <div class="watched-thread">
      <div class="bump-count">{item.bumps}</div>
      <div class="thread-link">
        <a href="/t/{b36encodenumber(item.id)}">
          {item.topic ?? `#${b36encodenumber(item.id)}`}
        </a>
      </div>
    </div>
  {/each}
{/if}
