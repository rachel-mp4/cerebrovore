<script lang="ts">
  import { b36encodenumber } from "../utils";
  import type { WatchThread } from "../types";
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
</script>

{#if items.length === 0}
  <div>nothing so far...</div>
{:else if usingBumpOrdering}
  {#each bumpsOrder as item}
    <div class="watched-thread">
      <div class="bump-count">{b36encodenumber(item.bumps)}</div>
      <div class="thread-link">
        <a href="/t/{b36encodenumber(item.id)}">
          {item.topic ?? `#${b36encodenumber(item.id)}`}
        </a>
      </div>
    </div>
  {/each}
{:else}
  {#each bumpOrder as item}
    <div class="watched-thread">
      <div class="bump-count">{b36encodenumber(item.bumps)}</div>
      <div class="thread-link">
        <a href="/t/{b36encodenumber(item.id)}">
          {item.topic ?? `#${b36encodenumber(item.id)}`}
        </a>
      </div>
    </div>
  {/each}
{/if}
