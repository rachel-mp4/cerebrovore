<script lang="ts">
  import type { Message } from "../types";
  import diff from "fast-diff";
  interface Props {
    message: Message;
    mylocaltext?: string;
  }
  let { message, mylocaltext }: Props = $props();

  let diffs = $derived(
    !message.lrcdata.pub && message.lrcdata.mine && mylocaltext
      ? diff(message.lrcdata.body, mylocaltext)
      : null,
  );
</script>

<div>
  {#if diffs}
    {#each diffs as diff}
      {#if diff[0] === -1}
        <span class="removed">{diff[1]}</span>
      {:else if diff[0] === 0}
        <span>{diff[1]}</span>
      {:else}
        <span class="appended">{diff[1]}</span>
      {/if}
    {/each}
  {:else if message.renderedHTML}
    {@html message.renderedHTML}
  {:else}
    {message.lrcdata.body}
  {/if}
</div>
