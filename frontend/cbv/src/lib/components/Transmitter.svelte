<script lang="ts">
  import AutoGrowInput from "./AutoGrowInput.svelte";
  import { WSContext } from "../wscontext.svelte";
  import { numToHex } from "../colors";
  import diff from "fast-diff";
  interface Props {
    ctx: WSContext;
    defaultNick?: string;
  }
  let { ctx }: Props = $props();
  let nick = $state("wanderer");
  let imageURL: string | undefined = $state();
  let imageAlt: string = $state("");
  let image: HTMLImageElement | undefined = $state();
  $effect(() => {
    if (ctx) {
      ctx.setNick(nick);
    }
  });
  const setName = (event: Event) => {
    const el = event.target as HTMLInputElement;
    ctx.nick = el.value;
  };

  let message = $state("");
  const addReply = (str: string) => {
    message = message + `${str}\n`;
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
    const text = reply.textContent;
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
        return;
      }
      inputEl.style.height = "auto";
      if (ctx.myMessage === undefined) {
        event.preventDefault();
        return;
      }
      ctx.insertLineBreak();
      message = "";
      event.preventDefault();
      sentmessage = "";
      return;
    }
  };

  const convertFileToImageItem = (blob: File) => {
    cancelimagepost();
    const blobUrl = URL.createObjectURL(blob);
    ctx.initImage(blob);
    imageURL = blobUrl;
  };
  const cancelimagepost = () => {
    if (imageURL) {
      URL.revokeObjectURL(imageURL);
    }
    ctx.cancelImage();
    imageAlt = "";
    imageURL = undefined;
  };
  const uploadimage = () => {
    ctx.pubImage(imageAlt);
    if (imageURL) {
      URL.revokeObjectURL(imageURL);
    }
    imageAlt = "";
    imageURL = undefined;
  };
  let inputEl: HTMLTextAreaElement;
  function adjustHeight() {
    if (inputEl) {
      const init = inputEl.style.height;
      inputEl.style.height = inputEl.scrollHeight + "px";
      if (inputEl.style.height !== init) {
        const el: HTMLElement | null = document.getElementById("main-content");
        if (el) {
          console.log("scrolling here!");
          setTimeout(() => {
            el.scrollTo(0, el.scrollHeight + 1000);
          }, 0);
        }
      }
    }
  }
  function adjust(event: Event) {
    diffAndSend();
  }
  $effect(() => {
    message;
    // we want to adjust height always, unless both the user has a
    // blank message and there is a hash in the url (in that case, )
    if (message === "" && location.hash !== "") {
      return;
    }
    adjustHeight();
  });
  const pastifier = (event: ClipboardEvent) => {
    const items = event.clipboardData?.items;
    if (items === undefined) {
      return;
    }
    for (const item of items) {
      if (item.type.startsWith("image/")) {
        const blob = item.getAsFile();
        if (blob === null) {
          return;
        }
        event.preventDefault();
        convertFileToImageItem(blob);
      }
    }
  };
</script>

<div class="transmitter">
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
      onInput={setName}
      maxlength={12}
      bold={true}
    />
    <input type="checkbox" name="anon" />
    <label for="anon">anon</label>
  </div>
  <div class="autogrowwrapper">
    <textarea
      rows="1"
      bind:value={message}
      bind:this={inputEl}
      maxlength={65535}
      oninput={adjust}
      onpaste={pastifier}
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
        class="media-upload"
        type="file"
        accept="image/*"
      />
    {/if}
  </div>
  {#if imageURL !== undefined}
    <div>
      <img bind:this={image} src={imageURL} alt={imageAlt} />
      <AutoGrowInput
        bind:value={imageAlt}
        placeholder="alt text"
        size={10}
        bold={false}
      />
      <button onclick={cancelimagepost}> cancel </button>
      {#if ctx.myImageRef !== undefined}
        <button onclick={uploadimage}> confirm </button>
      {:else}
        uploading...
      {/if}
    </div>
  {/if}
</div>
