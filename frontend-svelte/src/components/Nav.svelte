<script lang="ts">
  import { _, locale } from "svelte-i18n";
  import { user } from "../store";

  export let segment: string;
  let links = [
    { segment: undefined, href: ".", name: "home" },
    { segment: "candidates", href: "candidates", name: "candidates" },
    { segment: "faq", href: "faq", name: "faq" },
  ];

  async function logout() {
    await fetch("/api/auth/logout");
    user.set(null);
    window.location.href = "/";
  }

  function setLanguage(lang: string) {
    // TODO save locale preference in browser, and retrieve it on init...
    locale.set(lang);
  }
</script>

<style>
  nav.navbar {
    margin-bottom: 10px;
  }
</style>

<nav class="navbar navbar-expand-lg sticky-top navbar-dark bg-primary">
  <span class="navbar-brand">Bella Ciao</span>
  <button
    class="navbar-toggler"
    type="button"
    data-toggle="collapse"
    data-target="#navbar"
    aria-controls="navbar"
    aria-expanded="false"
    aria-label="Toggle navigation">
    <span class="navbar-toggler-icon" />
  </button>
  <div class="collapse navbar-collapse" id="navbar">
    <ul class="navbar-nav mr-auto">
      {#each links as link}
        <li class="nav-item" class:active={segment === link.segment}>
          <a
            class="nav-link"
            aria-current={segment === link.segment ? 'page' : undefined}
            href={link.href}>
            {$_("comp.nav."+link.name)}
            {#if segment === link.segment}
              <span class="sr-only">(current)</span>
            {/if}
          </a>
        </li>
      {/each}
      <li class="nav-item dropdown">
        <a
          class="nav-link dropdown-toggle"
          href="."
          id="languagesDropdown"
          role="button"
          data-toggle="dropdown"
          aria-haspopup="true"
          aria-expanded="false">
          {$_('comp.nav.language')}
        </a>
        <div class="dropdown-menu" aria-labelledby="languagesDropdown">
          <a
            class="dropdown-item"
            href="."
            on:click={() => setLanguage('ca')}>Català</a>
          <a
            class="dropdown-item"
            href="."
            on:click={() => setLanguage('en')}>English</a>
        </div>
      </li>
    </ul>
    {#if $user}
      <span class="navbar-text">
        <a class="nav-link" href="." on:click={logout}>{$_("comp.nav.logout")}</a>
      </span>
    {/if}
  </div>
</nav>
