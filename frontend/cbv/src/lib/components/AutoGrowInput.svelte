<script lang="ts">
  import { onMount } from "svelte";
  import { hexToTransparent } from "../colors";

  interface Props {
    onBeforeInput?: (event: InputEvent) => void;
    onInput?: (event: Event) => void;
    placeholder?: string;
    value?: string;
    size?: number;
    maxlength?: number;
    bold?: boolean;
    color?: string;
    fs?: string;
  }

  let {
    onBeforeInput,
    placeholder,
    value = $bindable(""),
    onInput,
    maxlength,
    size = 5,
    bold = false,
    color,
    fs,
  }: Props = $props();

  let inputEl: HTMLElement;
  function adjust(event: Event) {
    onInput?.(event);
  }

  function adjustWidth() {
    if (inputEl) {
      inputEl.style.width = "auto";
      inputEl.style.width = inputEl.scrollWidth + 1 + "px";
    }
  }
  onMount(adjustWidth);
  $effect(() => {
    value;
    adjustWidth();
  });
</script>

<input
  bind:value
  bind:this={inputEl}
  {maxlength}
  {size}
  oninput={adjust}
  onbeforeinput={onBeforeInput}
  style:font-weight={bold ? "700" : "inherit"}
  style:--theme={color}
  style:color={color ? "var(--theme)" : "var(--fg)"}
  style:--themep={hexToTransparent(color ?? "#ffffff")}
  style:font-size={fs ?? "1rem"}
  {placeholder}
/>

<style>
  input {
    background-color: transparent;
    border: none;
  }
  input::placeholder {
    color: var(--themep);
  }
</style>
