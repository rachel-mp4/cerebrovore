<script lang="ts">
  import type { Message } from "../types";
  import * as linkify from "linkifyjs";
  import { colorSetFromTheme } from "../colors";
  import type { ColorSet } from "../colors";
  import diff from "fast-diff";
  interface Props {
    message: Message;
    margin: number;
    mylocaltext?: string;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
    fs?: string;
  }
  let { message, margin, mylocaltext, onmute, onunmute, fs }: Props = $props();
  let nick: string | undefined = $derived(message.lrcdata.init?.nick);
  let handle: string | undefined = $derived(message.lrcdata.init?.handle);
  let color: number | undefined = $derived(message.lrcdata.init?.color);
  let colorSet: ColorSet = $derived(colorSetFromTheme(color ?? 8421504));
  const escapeHTML = (text: string): string => {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  };
  const convertLinksToMessageFrags = (body: string) => {
    const ebody = escapeHTML(body);
    const links = linkify.find(body, "url");
    const ll = links.length;
    if (ll === 0) {
      return [{ text: ebody, isLink: false, href: "", key: 0 }];
    }
    let res = [];
    let idx = 0;
    links.forEach((link) => {
      if (link.start > idx) {
        const beforeText = body.substring(idx, link.start);
        res.push({
          text: escapeHTML(beforeText),
          isLink: false,
          href: "",
          key: res.length,
        });
      }
      res.push({
        text: link.value,
        href: link.href,
        isLink: true,
        key: res.length,
      });
      idx = link.end;
    });
    if (idx < body.length) {
      const afterText = body.substring(idx);
      res.push({
        text: escapeHTML(afterText),
        isLink: false,
        key: res.length,
        href: "",
      });
    }
    return res;
  };
  let mfrags = $derived(convertLinksToMessageFrags(message.lrcdata.body));
  let diffs = $derived(
    message.lrcdata.pub && message.lrcdata.mine && mylocaltext
      ? diff(message.lrcdata.body, mylocaltext)
      : null,
  );
  let pinned = $state(false);
</script>

{#if message.lrcdata.muted === false}
  <div
    style:--theme={colorSet.theme}
    style:--themep={colorSet.themetransparent}
    style:--tcontrast={colorSet.themecontrast}
    style:--tpartial={colorSet.themecontrasttransparent}
    style:--margin={margin + "rem"}
    style:--size={fs ?? "1rem"}
    class="{message.lrcdata.pub ? '' : 'active'} 
    {pinned ? 'pinned' : ''} 
    {message.lrcdata.init ? '' : 'late'} 
    transmission"
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
        {#if message.lrcdata.mine !== true}
          <button
            class="mute clickable"
            onclick={() => {
              message.lrcdata.muted = true;
              onmute?.(message.id);
            }}
          >
            mute
          </button>
        {/if}
      {/if}
    </div>
    <div class="body">
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
      {:else}
        {#each mfrags as part (part.key)}
          {#if part.isLink}
            <a href={part.href} target="_blank" rel="noopener">{part.text}</a>
          {:else}
            {@html part.text}
          {/if}
        {/each}
      {/if}
    </div>
  </div>
{:else}
  muted.
  <button
    class="unmute"
    onclick={() => {
      message.lrcdata.muted = false;
      onunmute?.(message.id);
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
  .removed {
    text-decoration: line-through var(--tpartial);
  }
  .appended {
    color: var(--tpartial);
  }
  .transmission:not(:hover) .clickable {
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
  .active a {
    color: var(--tcontrast);
  }
  .handle {
    color: var(--fl);
    position: relative;
  }

  .nick {
    white-space: pre-wrap;
  }

  .transmission {
    padding-bottom: 1rem;
    margin-top: var(--margin);
    font-size: var(--size);
  }

  .transmission:not(.active) .header {
    color: var(--theme);
  }
  .body {
    white-space: pre-wrap;
    word-wrap: break-word;
    max-width: 100%;
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
</style>
