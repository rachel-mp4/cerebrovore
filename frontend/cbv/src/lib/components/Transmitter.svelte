<script lang="ts">
  import AutoGrowInput from "./AutoGrowInput.svelte";
  import { WSContext } from "../wscontext.svelte";
  import { numToHex } from "../colors";
  import diff from "fast-diff";
  import { onMount } from "svelte";
  interface Props {
    ctx: WSContext;
    defaultnick: string | null;
    defaultcolor: number | null;
    convertFileToImageItem: (blob: File) => void;
  }
  let { ctx, defaultnick, defaultcolor, convertFileToImageItem }: Props =
    $props();
  let nick = $state(ctx.nick);
  let anon = $state(ctx.anon);
  onMount(() => {
    if (window.location.hash === "") {
      document.dispatchEvent(new CustomEvent("lrc:scroll"));
    }
  });
  $effect(() => {
    if (ctx) {
      ctx.setNick(nick);
    }
  });

  $effect(() => {
    if (ctx) {
      ctx.setAnon(anon);
      setnameandcolor();
    }
  });

  var initial = true;
  const setnameandcolor = () => {
    if (initial) {
      initial = false;
      return;
    }
    if (anon) {
      nick = "wanderer";
      ctx.setColor(4534186);
    } else {
      if (defaultnick !== null) {
        nick = defaultnick;
      }
      if (defaultcolor !== null) {
        ctx.setColor(defaultcolor);
      }
    }
  };

  let message = $state("");
  const addReply = (str: string) => {
    const das = message !== "";
    const enl = message.endsWith("\n");
    // if it IS empty or if it ends with a new line, add a new line, otherwise a space
    message = message + `${str}${!das || enl ? "\n" : " "}`;
    // don't send our first reply
    if (das) {
      diffAndSend();
    }
  };
  document.addEventListener("click", (e: MouseEvent) => {
    const t = e.target as HTMLElement;
    if (!t) {
      return;
    }
    const reply = t.closest("button.reply");
    if (!reply) {
      return;
    }
    const text = reply.textContent?.trim();
    if (!text) {
      return;
    }
    addReply(text);
    inputEl.focus();
  });

  let color = $derived(numToHex(ctx.color));
  let sentmessage = $state("");
  const diffAndSend = () => {
    const result = diff(sentmessage, message);
    let idx = 0;
    result.forEach((d) => {
      switch (d[0]) {
        case -1:
          const idx2 = idx + d[1].length;
          ctx.delete(idx, idx2);
          break;
        case 0:
          idx = idx + d[1].length;
          break;
        case 1:
          ctx.insert(idx, d[1]);
          idx = idx + d[1].length;
          break;
      }
    });
    sentmessage = message;
  };

  const bi = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
      if (event.shiftKey) {
        event.stopPropagation();
        return;
      }
      event.preventDefault();
      const sent = ctx.insertLineBreak();
      switch (sent) {
        case 0: {
          // we don't stopPropagation here because we want to then
          // try and handle the event by uploading an image that is
          // ready to upload
          return;
        }

        case 1: {
          // normal case
          event.stopPropagation();
          message = "";
          sentmessage = "";
          return;
        }
        case 2: {
          // pubbedWithoutRecieving case
          event.stopPropagation();
          inputEl.classList.add("dzzzt");
          dezzzt += 1;
          setTimeout(() => {
            dezzzt -= 1;
            if (dezzzt === 0) {
              inputEl.classList.remove("dzzzt");
            }
          }, 340);
        }
      }
    }
  };
  var dezzzt = 0;

  let lines = $state(1);
  let inputEl: HTMLTextAreaElement;
  function adjustHeight() {
    if (inputEl) {
      inputEl.style.height = "auto";
      const pheight = Math.min(inputEl.scrollHeight, 24 * 4.5);
      lines = Math.round(pheight / 12) / 2; // presumably 9 / 2 always equals 4.5 in floating point >_>
      inputEl.style.height = `${pheight}px`;
    }
  }
  function adjust(event: Event) {
    diffAndSend();
  }
  $effect(() => {
    message;
    adjustHeight();
  });
  onMount(() => {
    inputEl.focus();
  });
</script>

{#if ctx.systemMessage}
  <div class="system-message">{ctx.systemMessage}</div>
{/if}
<div id="transmitter-thingy" data-lines={lines} class="transmitter">
  <div class="wrapper" style:--accent={color}>
    <input
      type="range"
      min="0"
      max="16777215"
      bind:value={ctx.color}
      onchange={() => {
        ctx.setColor(ctx.color);
      }}
    />
    <AutoGrowInput
      bind:value={nick}
      {color}
      size={4}
      placeholder="alice!"
      maxlength={16}
      bold={true}
    />
    <input type="checkbox" name="anon-box" id="anon-box" bind:checked={anon} />
    <label for="anon-box">anon</label>
    {#if ctx.rttping}<span>{ctx.rttping}</span>{/if}
  </div>
  <div class="autogrowwrapper">
    <textarea
      rows="1"
      bind:value={message}
      bind:this={inputEl}
      maxlength={65535}
      oninput={adjust}
      onkeydown={bi}
      placeholder="start typing..."
    ></textarea>
    {#if message === ""}
      <label class="media-upload-button" for="media-upload"
        >...or upload an image</label
      >
      <input
        onchange={(event) => {
          if (
            event.currentTarget.files &&
            event.currentTarget.files.length > 0
          ) {
            convertFileToImageItem(event.currentTarget.files[0]);
          }
        }}
        id="media-upload"
        type="file"
        accept="image/*"
      />
    {/if}
  </div>
</div>
