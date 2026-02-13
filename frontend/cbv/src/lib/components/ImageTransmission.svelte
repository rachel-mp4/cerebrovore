<script lang="ts">
  import type { Image } from "../types";
  import { colorSetFromTheme, type ColorSet } from "../colors";
  import { smartAbsoluteTimestamp, dumbAbsoluteTimestamp } from "../utils";
  interface Props {
    image: Image;
    margin: number;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
    fs?: string;
  }
  let { image, margin, onmute, onunmute, fs }: Props = $props();
  let nick: string | undefined = $derived(image.lrcdata.init?.nick);
  let handle: string | undefined = $derived(image.lrcdata.init?.handle);
  let color: number | undefined = $derived(image.lrcdata.init?.color);
  let colorSet: ColorSet = $derived(colorSetFromTheme(color ?? 8421504));
  let src: string | undefined = $derived(image.lrcdata.pub?.contentAddress);
  let alt: string | undefined = $derived(image.lrcdata.pub?.alt);
  let pinned = $state(false);
</script>

{#if image.lrcdata.muted === false}
  <div
    style:--theme={colorSet.theme}
    style:--themep={colorSet.themetransparent}
    style:--tcontrast={colorSet.themecontrast}
    style:--tpartial={colorSet.themecontrasttransparent}
    style:--margin={margin + "rem"}
    style:--size={fs ?? "1rem"}
    class="{image.lrcdata.pub ? '' : 'active'} 
    {pinned ? 'pinned' : ''} 
    {image.lrcdata.init ? '' : 'late'} 
    imageTransmission"
  >
    <div class="header">
      <span class="nick">{nick ?? ""}</span>{#if handle !== undefined}
        <span class="handle">@{handle}</span>
        <button
          class="clickable"
          onclick={() => {
            pinned = !pinned;
          }}
        >
          {pinned ? "unpin" : "pin"}
        </button>
        {#if image.lrcdata.mine !== true}
          <button
            class="mute clickable"
            onclick={() => {
              image.lrcdata.mine = true;
              onmute?.(image.id);
            }}
          >
            mute
          </button>
        {/if}
      {/if}
    </div>
    {#if src}
      <div class="image-wrapper">
        <img class="bg-img" {src} {alt} />
        <img class="fg-img" {src} {alt} />
      </div>
    {:else}
      i don't have an image yet
    {/if}
  </div>
{:else}
  muted.
  <button
    class="unmute"
    onclick={() => {
      image.lrcdata.muted = false;
      onunmute?.(image.id);
    }}
  >
    unmute
  </button>
{/if}

<style>
  .active {
    position: relative;
    background-color: var(--themep);
    color: var(--tcontrast);
  }
  .active::before {
    position: absolute;
    content: "";
    inset: 0;
    z-index: -1;
    background-color: var(--theme);
  }
  .imageTransmission:not(:hover) .clickable {
    display: none;
  }
  .active .clickable {
    color: var(--tpartial);
  }
  .clickable {
    color: var(--fl);
    cursor: pointer;
  }
  .clickable:hover {
    color: var(--fg);
  }
  .active .clickable:hover {
    color: var(--contrast);
  }
  .pinned {
    order: 1;
  }

  .header {
    font-weight: 700;
  }
  .active .handle {
    color: var(--tpartial);
  }
  .handle {
    color: var(--fl);
    position: relative;
  }

  .nick {
    white-space: pre-wrap;
  }

  .handle::after {
    content: "";
    color: var(--fg);
    background: var(--bg);
    position: absolute;
    left: 0;
    right: 0;
    top: calc(55% - calc(var(--size) / 8));
    bottom: calc(45% - calc(var(--size) / 8));
    transform: scaleX(0);
    transform-origin: center left;
    transition: transform 0.17s 3s;
  }

  .imageTransmission:not(.signed):not(.active) .handle::after {
    transform: scaleX(1);
  }
  .imageTransmission:not(.signed):not(.active) .handle:hover::after {
    content: "i couldn't find a record :c";
  }

  .imageTransmission {
    padding-bottom: 1rem;
    margin-top: var(--margin);
    font-size: var(--size);
  }

  .imageTransmission:not(.active) .header {
    color: var(--theme);
  }
  button {
    font-size: var(--size);
    background-color: transparent;
    border: none;
    color: var(--fg);
    padding: 0;
    cursor: pointer;
  }
  button:hover {
    font-weight: 700;
  }
  .image-wrapper {
    position: relative;
  }
  .image-wrapper .bg-img {
    position: absolute;
    z-index: -1;
  }
  .image-wrapper .fg-img {
    position: relative;
    opacity: 0.5;
  }
</style>
