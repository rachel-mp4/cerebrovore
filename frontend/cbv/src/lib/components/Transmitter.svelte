<script lang="ts">
  import AutoGrowInput from "./AutoGrowInput.svelte";
  import { WSContext } from "../wscontext.svelte";
  import { numToHex } from "../colors";
  import AutoGrowTextArea from "./AutoGrowTextArea.svelte";
  import diff from "fast-diff";
  interface Props {
    ctx: WSContext;
    defaultNick?: string;
    defaultHandle?: string;
  }
  let { ctx, defaultNick, defaultHandle }: Props = $props();
  let nick = $state(defaultNick ?? "wanderer");
  let innerWidth = $state(0);
  let isDesktop = $derived(innerWidth > 1000);
  let imageURL: string | undefined = $state();
  let imageAlt: string = $state("");
  let image: HTMLImageElement | undefined = $state();
  $effect(() => {
    if (ctx) {
      ctx.setNick(nick);
    }
  });
  let message = $state("");
  const setName = (event: Event) => {
    const el = event.target as HTMLInputElement;
    ctx.nick = el.value;
  };
  let handle = $state(defaultHandle ?? "cool.org");
  const setHandle = (event: Event) => {
    const el = event.target as HTMLInputElement;
    ctx.handle = el.value;
  };
  $effect(() => {
    if (ctx) {
      ctx.setHandle(handle);
    }
  });

  let color = $derived(numToHex(ctx.color));
  const diffAndSendEl = (el: HTMLTextAreaElement) => {
    const result = diff(message, el.value);
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
    message = el.value;
  };

  const bi = (event: InputEvent) => {
    const el = event.target as HTMLInputElement;
    switch (event.inputType) {
      case "insertLineBreak": {
        if (ctx.myMessage === undefined) {
          event.preventDefault();
          return;
        }
        ctx.insertLineBreak();
        el.value = "";
        event.preventDefault();
        message = "";
        return;
      }
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
</script>

<svelte:window bind:innerWidth />

<div id="transmitter">
  <div
    class="wrapper"
    style:--theme={color}
    style:font-size={isDesktop ? "2rem" : "1rem"}
  >
    <input
      type="range"
      min="0"
      max="16777215"
      bind:value={ctx.color}
      onchange={() => {
        ctx.setColor(ctx.color);
      }}
      style:font-size={isDesktop ? "2rem" : "1rem"}
    />
    <AutoGrowInput
      bind:value={nick}
      {color}
      size={4}
      placeholder="alice!"
      onInput={setName}
      maxlength={12}
      bold={true}
      fs={isDesktop ? "2rem" : "1rem"}
    />@<AutoGrowInput
      bind:value={handle}
      placeholder="alice.com"
      size={8}
      onInput={setHandle}
      maxlength={253}
      bold={false}
      fs={isDesktop ? "2rem" : "1rem"}
    />
  </div>
  <AutoGrowTextArea
    placeholder="start typing..."
    onBeforeInput={bi}
    onInputEl={diffAndSendEl}
    imageHandler={convertFileToImageItem}
    maxlength={65535}
    fs={isDesktop ? "2rem" : "1rem"}
  />
  {#if imageURL !== undefined}
    <div>
      <img bind:this={image} src={imageURL} alt={imageAlt} />
      <AutoGrowInput
        bind:value={imageAlt}
        placeholder="alt text"
        size={10}
        bold={false}
        fs={isDesktop ? "2rem" : "1rem"}
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

<style>
  #transmitter {
    flex-shrink: 0;
  }
  .wrapper {
    position: relative;
    display: inline-block;
  }
  .wrapper :first-child {
    position: absolute;
    left: 0px;
    right: 0px;
    top: -2rem;
    bottom: 1em;
    display: none;
    accent-color: var(--theme);
  }
  .wrapper:focus-within :first-child {
    display: inline-block;
  }
</style>
