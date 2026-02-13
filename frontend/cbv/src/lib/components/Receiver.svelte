<script lang="ts">
  import EnbyTransmission from "./EnbyTransmission.svelte";
  import MessageTransmission from "./MessageTransmission.svelte";
  import ImageTransmission from "./ImageTransmission.svelte";
  import type { Item } from "../types";
  import { isMessage, isImage, isEnby } from "../types";
  interface Props {
    items: Array<Item>;
    mylocaltext?: string;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
  }
  let { items, mylocaltext, onmute, onunmute }: Props = $props();
  let length = $derived(items.length);
  let innerWidth = $state(0);
  let isDesktop = $derived(innerWidth > 1000);
</script>

<svelte:window bind:innerWidth />
<div id="receiver">
  {#each items as item, index (`${item.id}-${item.type}`)}
    {@const last = length - 1}
    {@const diff = last - index}
    {@const guess = 2 + (Math.atan((diff - 19) * 0.25) / -2.8 - 0.45)}
    {@const res = Math.min(Math.max(guess, 1), 2)}
    {#if isEnby(item)}
      <EnbyTransmission enby={item} />
    {:else if isMessage(item)}
      <MessageTransmission
        message={item}
        mylocaltext={item.lrcdata.mine && !item.lrcdata.pub
          ? mylocaltext
          : undefined}
        margin={0}
        {onmute}
        {onunmute}
        fs={isDesktop ? `${res}rem` : "1rem"}
      />
    {:else if isImage(item)}
      <ImageTransmission
        image={item}
        margin={0}
        {onmute}
        {onunmute}
        fs={isDesktop ? `${res}rem` : "1rem"}
      />
    {/if}
  {/each}
</div>

<style>
  #receiver {
    width: 100%;
    display: flex;
    flex-direction: column;
    flex: 1;
  }
</style>
