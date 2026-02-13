<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    onBeforeInput?: (event: InputEvent) => void;
    onInputEl?: (el: HTMLTextAreaElement) => void;
    imageHandler?: (image: File) => void;
    placeholder?: string;
    value?: string;
    maxlength?: number;
    bold?: boolean;
    color?: string;
    fs?: string;
  }

  let {
    onBeforeInput,
    placeholder,
    value = $bindable(""),
    onInputEl,
    imageHandler,
    maxlength,
    bold = false,
    color,
    fs,
  }: Props = $props();

  let inputEl: HTMLTextAreaElement;
  function adjust(event: Event) {
    onInputEl?.(inputEl);
  }

  function bi(event: InputEvent) {
    onBeforeInput?.(event);
    adjustHeight();
  }

  function adjustHeight() {
    if (inputEl) {
      inputEl.style.height = "auto";
      inputEl.style.height = inputEl.scrollHeight + "px";
    }
  }
  onMount(adjustHeight);
  $effect(() => {
    value;
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
        imageHandler?.(blob);
      }
    }
  };
</script>

<div class="autogrowwrapper" style:--fs={fs ?? "1rem"}>
  <textarea
    rows="1"
    bind:value
    bind:this={inputEl}
    {maxlength}
    oninput={adjust}
    onpaste={pastifier}
    onbeforeinput={bi}
    style:font-weight={bold ? "700" : "inherit"}
    style:--theme={color}
    style:font-size="var(--fs)"
    {placeholder}
  ></textarea>
  {#if value === ""}
    <label class="media-upload-button" for="media-upload"
      >...or upload an image</label
    >
    <input
      onchange={(event) => {
        if (event.currentTarget.files && event.currentTarget.files.length > 0) {
          imageHandler?.(event.currentTarget.files[0]);
        }
      }}
      id="media-upload"
      type="file"
      accept="image/*"
    />
  {/if}
</div>

<style>
  #media-upload {
    display: none;
  }
  .media-upload-button {
    position: absolute;
    top: 0;
    right: 0;
    margin: 0;
    font-size: var(--fs);
    border: none;
    color: var(--bl);
    cursor: pointer;
  }
  .media-upload-button:hover {
    font-weight: 700;
  }

  textarea {
    width: 100%;
    font: inherit;
    padding: 0;
    margin: 0;
    display: block;
    resize: none;
    border: none;
    background: var(--fl);
  }
  .autogrowwrapper {
    position: relative;
  }

  .autogrowwrapper::before {
    content: "";
    position: absolute;
    inset: 0;
    background: var(--fg);
    z-index: -1;
  }
</style>
