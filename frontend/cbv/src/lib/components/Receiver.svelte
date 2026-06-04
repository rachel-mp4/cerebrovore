<script lang="ts">
  import { onMount } from "svelte";
  import { b36encodenumber } from "../utils";
  import EnbyTransmission from "./EnbyTransmission.svelte";
  import MessageTransmission from "./MessageTransmission.svelte";
  import ImageTransmission from "./ImageTransmission.svelte";
  import AutoGrowInput from "./AutoGrowInput.svelte";
  import type { Item } from "../types";
  import { newAbsoluteTimestamp, maxItemIdx } from "../utils";
  import { isMessage, isImage, isEnby } from "../types";
  import { numIsDark, numToHex, hexToTransparent } from "../colors";
  import { tick } from "svelte";
  import type { WSContext } from "../wscontext.svelte";
  interface Props {
    items: Array<Item>;
    mylocalid?: string;
    mylocaltext?: string;
    mylocalimage?: string | undefined;
    onmute?: (id: number) => void;
    onunmute?: (id: number) => void;
    ismoderator: boolean;
    cancelimagepost: () => void;
    uploadimage: (alt: string | undefined) => void;
    ctx: WSContext;
  }
  let {
    items,
    mylocalid,
    mylocaltext,
    mylocalimage,
    onmute,
    onunmute,
    ismoderator,
    cancelimagepost,
    uploadimage,
    ctx,
  }: Props = $props();
  let bottomEl: HTMLDivElement;
  let pinnedToBottomone = true;
  let pinnedToBottomtwo = true;
  let pinnedToBottomthree = true;
  let pinnedToBottomfour = true;
  let pinnedToBottomfourpfive = true;
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

  document.addEventListener("lrc:append", async () => {
    await tick();
    const el = document.getElementById("eats-ur-brain");
    if (el === null) {
      return;
    }
    var child: Element | null = el.lastElementChild;
    if (child === null) {
      return;
    }
    if (child !== null) {
      while (!child.classList.contains("tx")) {
        const prevchild: Element | null = child.previousElementSibling;
        if (prevchild === null) {
          return;
        }
        child = prevchild;
      }
      for (let i = 1; i < 40; i++) {
        child.setAttribute("data-from-end", i.toString());
        const prevchild: Element | null = child.previousElementSibling;
        if (prevchild === null) {
          return;
        }
        child = prevchild;
      }
    }
  });

  let pendingscroll = false;
  const scrollIfPinned = () => {
    const lines = document
      .getElementById("transmitter-thingy")
      ?.getAttribute("data-lines");
    switch (lines) {
      case "1":
        if (pinnedToBottomone) scroll();
        break;
      case "2":
        if (pinnedToBottomtwo) scroll();
        break;
      case "3":
        if (pinnedToBottomthree) scroll();
        break;
      case "4":
        if (pinnedToBottomfour) scroll();
        break;
      case "4.5":
        if (pinnedToBottomfourpfive) scroll();
        break;
      default:
        console.log("i wanna die");
    }
  };

  const scroll = () => {
    if (pendingscroll) return;
    pendingscroll = true;
    requestAnimationFrame(() => {
      pendingscroll = false;
      bottomEl.scrollIntoView();
    });
  };

  onMount(() => {
    const observerone = new IntersectionObserver(
      ([entry]) => {
        pinnedToBottomone = entry.isIntersecting;
      },
      {
        rootMargin: `0px 0px -${24 + 24 * 1}px 0px`,
      },
    );
    const observertwo = new IntersectionObserver(
      ([entry]) => {
        pinnedToBottomtwo = entry.isIntersecting;
      },
      {
        rootMargin: `0px 0px -${24 + 24 * 2}px 0px`,
      },
    );
    const observerthree = new IntersectionObserver(
      ([entry]) => {
        pinnedToBottomthree = entry.isIntersecting;
      },
      {
        rootMargin: `0px 0px -${24 + 24 * 3}px 0px`,
      },
    );
    const observerfour = new IntersectionObserver(
      ([entry]) => {
        pinnedToBottomfour = entry.isIntersecting;
      },
      {
        rootMargin: `0px 0px -${24 + 24 * 4}px 0px`,
      },
    );
    const observerfourpfive = new IntersectionObserver(
      ([entry]) => {
        pinnedToBottomfourpfive = entry.isIntersecting;
      },
      {
        rootMargin: `0px 0px -${24 + 24 * 4.5}px 0px`,
      },
    );
    observerone.observe(bottomEl);
    observertwo.observe(bottomEl);
    observerthree.observe(bottomEl);
    observerfour.observe(bottomEl);
    observerfourpfive.observe(bottomEl);
    document.addEventListener("lrc:scroll", scroll);
    document.addEventListener("lrc:scrollIfAttached", scrollIfPinned);

    return () => {
      observerone.disconnect();
      observertwo.disconnect();
      observerthree.disconnect();
      observerfour.disconnect();
      observerfourpfive.disconnect();
      document.removeEventListener("lrc:scroll", scroll);
      document.removeEventListener("lrc:scrollIfAttached", scrollIfPinned);
    };
  });
  $effect(() => {
    if (mylocaltext && mylocalid) {
      scrollIfPinned();
    }
  });
  let alt: string = $state("");
</script>

{#each items as item (`${item.id}-${item.type}`)}
  {@const id = b36encodenumber(item.id)}
  {@const showid = item.id < maxItemIdx}
  {#if item.lrcdata.muted === false}
    <div
      {id}
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
        {#if showid}
          <button class="reply">{` #${id}`}</button>
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
              <a class="moderate clickable" href={`/moderate?id=${id}`}
                >{" moderate"}</a
              >
            {/if}
          {:else}
            <button class="delete clickable">{" delete"}</button>
            {#if ismoderator}
              <a class="moderate clickable" href={`/moderate?id=${id}`}
                >{" moderate"}</a
              >
            {/if}
          {/if}
        {/if}
        {#if item.pubAt !== undefined}
          <time class="time" datetime={new Date(item.pubAt).toISOString()}
            >{newAbsoluteTimestamp(item.pubAt)}</time
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
              {alt}
              gifoverride={true}
            />
            <div class="collapse">
              <div>(this is a preview that only you can see)</div>
              <div>
                <button
                  class="clickable"
                  onclick={() => {
                    cancelimagepost();
                    alt = "";
                  }}
                >
                  cancel
                </button>
                <AutoGrowInput
                  submit={() => {
                    if (ctx.myMediaUploadState.kind === "uploaded") {
                      uploadimage(alt);
                      alt = "";
                    }
                  }}
                  bind:value={alt}
                  placeholder="alt text (optional)"
                  size={10}
                  bold={false}
                  fs={"inherit"}
                />
                {#if ctx.myMediaUploadState.kind === "ready"}
                  something went wrong if you can see me tbh
                {:else if ctx.myMediaUploadState.kind === "uploading"}
                  uploading...
                {:else}
                  <button
                    onclick={() => {
                      uploadimage(alt === "" ? undefined : alt);
                      alt = "";
                    }}
                  >
                    confirm
                  </button>
                {/if}
              </div>
            </div>
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
<div bind:this={bottomEl} style="height:1px"></div>
