<script lang="ts">
  import { b36encodenumber } from "../utils";
  import EnbyTransmission from "./EnbyTransmission.svelte";
  import MessageTransmission from "./MessageTransmission.svelte";
  import ImageTransmission from "./ImageTransmission.svelte";
  import type { Item } from "../types";
  import { newAbsoluteTimestamp } from "../utils";
  import { isMessage, isImage, isEnby } from "../types";
  import { numIsDark, numToHex, hexToTransparent } from "../colors";
  interface Props {
    items: Array<Item>;
    mylocaltext?: string;
    mylocalimage?: string | undefined;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
    ismoderator: boolean;
  }
  let {
    items,
    mylocaltext,
    mylocalimage,
    onmute,
    onunmute,
    ismoderator,
  }: Props = $props();
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
  let el = document.getElementById("main-content");
  $effect(() => {
    if (mylocaltext) {
      checknscroll();
    }
  });
  const checknscroll = () => {
    // i just use a big value here (100) because i can't figure it out...
    // besides, if you're typing, you probably want to be at bottom
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - 100) {
      setTimeout(() => {
        if (el) el.scrollTo(0, el.scrollHeight);
      }, 0);
    }
  };
</script>

{#each items as item (`${item.id}-${item.type}`)}
  {#if item.lrcdata.muted === false}
    <div
      id={b36encodenumber(item.id)}
      style:--accent={numToHex(item.lrcdata?.init?.color ?? 0)}
      style:--accentl={hexToTransparent(
        numToHex(item.lrcdata?.init?.color ?? 0),
      )}
      class="tx{isActive(item) ? ' active' : ''}{item.lrcdata.init
        ? ''
        : ' late'}{numIsDark(item.lrcdata?.init?.color ?? 0)
        ? ' light'
        : ' dark'}{item.lrcdata.mine === true ? ' you' : ''}"
    >
      <div class="header">
        {#if item.lrcdata.init}
          {#if item.lrcdata.init.nick}
            <span class="nick">{item.lrcdata.init.nick}</span>
          {/if}
          {#if item.lrcdata.init.handle !== undefined}
            <span class="handle"
              ><a href="/profile/{item.lrcdata.init.handle}"
                >@{item.lrcdata.init.handle}</a
              ></span
            >
          {/if}
        {/if}
        <button class="reply">{` #${b36encodenumber(item.id)}`}</button>
        {#if item.lrcdata.mine !== true}
          <button class="report clickable">{" report"}</button>
          <button
            class="mute clickable"
            onclick={() => {
              item.lrcdata.muted = true;
              onmute?.(item.id);
            }}>{" mute"}</button
          >
          {#if ismoderator}
            <button class="delete clickable">{" delete"}</button>
            <a
              class="moderate clickable"
              href={`/moderate?id=${b36encodenumber(item.id)}`}>{" moderate"}</a
            >
          {/if}
        {:else}
          <button class="delete clickable">{" delete"}</button>
          {#if ismoderator}
            <a
              class="moderate clickable"
              href={`/moderate?id=${b36encodenumber(item.id)}`}>{" moderate"}</a
            >
          {/if}
        {/if}
        {#if item.pubAt !== undefined}
          <time class="time" datetime={new Date(item.pubAt).toISOString()}
            >{newAbsoluteTimestamp(item.pubAt)}</time
          >
        {:else if isMessage(item) && item.lrcdata.mine}
          <time class="time"
            >{item.lrcdata.body === mylocaltext ? "there" : "here"}</time
          >
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
          {:else if item.lrcdata.mine && !item.lrcdata.pub}
            <ImageTransmission
              src={mylocalimage ?? ""}
              alt={undefined}
              gifoverride={true}
            />
            <div>THIS IS A PREVIEW THAT ONLY YOU CAN SEE</div>
          {:else}
            i don't have image yet
          {/if}
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
