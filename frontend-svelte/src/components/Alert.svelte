<script lang="ts">
  export let type: string = "info";
  export let content: string;
</script>

<div class={'alert alert-' + type} role="alert">{content}</div>
