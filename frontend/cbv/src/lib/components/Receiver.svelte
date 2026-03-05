<script lang="ts">
  import { b36encodenumber } from "../utils";
  import EnbyTransmission from "./EnbyTransmission.svelte";
  import MessageTransmission from "./MessageTransmission.svelte";
  import ImageTransmission from "./ImageTransmission.svelte";
  import type { Item } from "../types";
  import { isMessage, isImage, isEnby } from "../types";
  import { numIsDark, numToHex } from "../colors";
  interface Props {
    items: Array<Item>;
    mylocaltext?: string;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
  }
  let { items, mylocaltext, onmute, onunmute }: Props = $props();
  const isActive = (item: Item): boolean => {
    if (isEnby(item)) {
      return true;
    } else if (isMessage(item)) {
      return !item.lrcdata.pub;
    } else if (isImage(item)) {
      return !item.lrcdata.pub;
    } else {
      return false;
    }
  };
</script>

{#each items as item (`${item.id}-${item.type}`)}
  {#if item.lrcdata.muted === false}
    <div
      id={b36encodenumber(item.id)}
      style:--accent={numToHex(item.lrcdata?.init?.color ?? 0)}
      class="transmission{isActive(item) ? ' active' : ''}{item.lrcdata.init
        ? ''
        : ' late'}{numIsDark(item.lrcdata?.init?.color ?? 0)
        ? ' light'
        : ' dark'}"
    >
      <div class="header">
        {#if item.lrcdata.init}
          {#if item.lrcdata.init.nick}
            <span class="nick">{item.lrcdata.init.nick}</span>
          {/if}
          {#if item.lrcdata.init.handle !== undefined}
            <span class="handle">@{item.lrcdata.init.handle}</span>
          {/if}
        {/if}
        <button class="reply">#{b36encodenumber(item.id)}</button>
        {#if item.lrcdata.mine !== true}
          <button
            class="mute clickable"
            onclick={() => {
              item.lrcdata.muted = true;
              onmute?.(item.id);
            }}
          >
            mute
          </button>
        {/if}
      </div>
      <div class="body">
        {#if isEnby(item)}
          <EnbyTransmission enby={item} />
        {:else if isMessage(item)}
          <MessageTransmission
            message={item}
            mylocaltext={item.lrcdata.mine && !item.lrcdata.pub
              ? mylocaltext
              : undefined}
          />
        {:else if isImage(item)}
          {#if item.lrcdata.pub?.contentAddress}
            <ImageTransmission
              src={item.lrcdata.pub.contentAddress}
              alt={item.lrcdata.pub.alt}
            />
          {:else}i don't have image yet{/if}
        {/if}
      </div>
      {#if item.replies !== null}
        <div class="footer">
          {#each item.replies as reply}
            <a href="/p/{b36encodenumber(reply)}">#{b36encodenumber(reply)}</a>
            {" "}
          {/each}
        </div>
      {/if}
    </div>
  {:else}
    muted
    <button
      class="unmute"
      onclick={() => {
        item.lrcdata.muted = false;
        onunmute?.(item.id);
      }}
    >
      unmute
    </button>
  {/if}
{/each}
