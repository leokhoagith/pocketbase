<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { scale, fly } from "svelte/transition";
    import { SchemaField } from "pocketbase";
    import CommonHelper from "@/utils/CommonHelper";
    import tooltip from "@/actions/tooltip";
    import { errors } from "@/stores/errors";
    import Accordion from "@/components/base/Accordion.svelte";
    import Field from "@/components/base/Field.svelte";
    import FieldTypeSelect from "@/components/collections/schema/FieldTypeSelect.svelte";
    import TextOptions from "@/components/collections/schema/TextOptions.svelte";
    import NumberOptions from "@/components/collections/schema/NumberOptions.svelte";
    import BoolOptions from "@/components/collections/schema/BoolOptions.svelte";
    import EmailOptions from "@/components/collections/schema/EmailOptions.svelte";
    import UrlOptions from "@/components/collections/schema/UrlOptions.svelte";
    import DateOptions from "@/components/collections/schema/DateOptions.svelte";
    import SelectOptions from "@/components/collections/schema/SelectOptions.svelte";
    import JsonOptions from "@/components/collections/schema/JsonOptions.svelte";
    import FileOptions from "@/components/collections/schema/FileOptions.svelte";
    import RelationOptions from "@/components/collections/schema/RelationOptions.svelte";
    import UserOptions from "@/components/collections/schema/UserOptions.svelte";

    const dispatch = createEventDispatcher();

    export let key = "0";
    export let field = new SchemaField();
    export let disabled = false;
    export let excludeNames = [];

    let accordion;
    let initialType = field.type;

    $: if (initialType != field.type) {
        initialType = field.type;
        // reset common options
        field.options = {};
        field.unique = false;
    }

    $: if (excludeNames.length) {
        const normalizedName = normalizeFieldName(field.name);
        if (field.name !== normalizedName) {
            field.name = normalizedName;
        }
    }

    $: canBeStored = !CommonHelper.isEmpty(field.name) && field.type;

    $: if (!canBeStored) {
        accordion && expand();
    }

    $: if (field.toDelete) {
        accordion && collapse();

        // reset the name if it was previously deleted
        if (!field.name && field.originalName) {
            field.name = field.originalName;
        }
    }

    $: if (!field.originalName && field.name) {
        field.originalName = field.name;
    }

    $: if (typeof field.toDelete === "undefined") {
        field.toDelete = false; // normalize
    }

    $: interactive = !disabled && !field.system && !field.toDelete && canBeStored;

    $: hasErrors = !CommonHelper.isEmpty(CommonHelper.getNestedVal($errors, `schema.${key}`));

    export function expand() {
        accordion?.expand();
    }

    export function collapse() {
        accordion?.collapse();
    }

    function handleDelete() {
        if (!field.id) {
            collapse();
            dispatch("remove");
        } else {
            field.toDelete = true;
        }
    }

    function normalizeFieldName(name) {
        name = CommonHelper.slugify(name);

        let counter = "";
        while (excludeNames.includes(name + counter)) {
            ++counter;
        }

        return name + counter;
    }

    onMount(() => {
        // auto expand new fields
        if (!field.id) {
            expand();
        }
    });
</script>

<Accordion
    bind:this={accordion}
    on:expand
    on:collapse
    on:toggle
    single
    {interactive}
    class={disabled || field.toDelete || field.system ? "field-accordion disabled" : "field-accordion"}
>
    <svelte:fragment slot="header" let:active={expanded}>
        <div class="inline-flex">
            <span class="icon field-type">
                <i class={CommonHelper.getFieldTypeIcon(field.type)} />
            </span>
            <strong class="title field-name" class:txt-strikethrough={field.toDelete} title={field.name}>
                {field.name || "-"}
            </strong>
        </div>

        {#if !field.toDelete}
            <div class="inline-flex">
                {#if field.system}
                    <span class="label label-danger">System</span>
                {/if}
                {#if !field.id}
                    <span class="label" class:label-warning={interactive && !field.toDelete}>New</span>
                {/if}
                {#if field.required}
                    <span class="label label-success">Required</span>
                {/if}
                {#if field.unique}
                    <span class="label label-success">Unique</span>
                {/if}
            </div>
        {/if}

        <div class="flex-fill" />

        {#if hasErrors}
            <i
                class="ri-error-warning-fill txt-danger"
                transition:scale={{ duration: 150, start: 0.7 }}
                use:tooltip={{ text: "Has errors", position: "left" }}
            />
        {/if}

        {#if expanded && !field.toDelete}
            <div class="inline-flex flex-gap-sm flex-nowrap" in:fly={{ duration: 200, x: 20, opacity: 0 }}>
                <button
                    type="button"
                    class="btn btn-sm fade p-l-0 p-r-0"
                    on:click|stopPropagation={handleDelete}
                >
                    <span class="txt">Remove</span>
                </button>

                {#if interactive}
                    <button
                        type="button"
                        class="btn btn-sm btn-outline btn-expanded-sm"
                        on:click|stopPropagation={collapse}
                    >
                        <span class="txt">Done</span>
                    </button>
                {/if}
            </div>
        {/if}

        {#if field.toDelete}
            <button
                type="button"
                class="btn btn-sm btn-danger btn-secondary"
                on:click|stopPropagation={() => {
                    field.toDelete = false;
                }}
            >
                <span class="txt">Restore</span>
            </button>
        {/if}
    </svelte:fragment>

    <form
        class="field-form"
        on:submit|preventDefault={() => {
            canBeStored && collapse();
        }}
    >
        <div class="grid">
            <div class="col-sm-6">
                <Field
                    class="form-field required {field.id ? 'disabled' : ''}"
                    name="schema.{key}.type"
                    let:uniqueId
                >
                    <label for={uniqueId}>Type</label>
                    <FieldTypeSelect id={uniqueId} disabled={field.id} bind:value={field.type} />
                </Field>
            </div>
            <div class="col-sm-6">
                <Field
                    class="form-field required {field.id && field.system ? 'disabled' : ''}"
                    name="schema.{key}.name"
                    let:uniqueId
                >
                    <label for={uniqueId}>Name</label>
                    <!-- svelte-ignore a11y-autofocus -->
                    <input
                        type="text"
                        id={uniqueId}
                        required
                        disabled={field.id && field.system}
                        spellcheck="false"
                        autofocus={!field.id}
                        value={field.name}
                        on:input={(e) => {
                            field.name = normalizeFieldName(e.target.value);
                            e.target.value = field.name;
                        }}
                    />
                </Field>
            </div>

            <div class="col-sm-12 hidden-empty">
                {#if field.type === "text"}
                    <TextOptions {key} bind:options={field.options} />
                {:else if field.type === "number"}
                    <NumberOptions {key} bind:options={field.options} />
                {:else if field.type === "bool"}
                    <BoolOptions {key} bind:options={field.options} />
                {:else if field.type === "email"}
                    <EmailOptions {key} bind:options={field.options} />
                {:else if field.type === "url"}
                    <UrlOptions {key} bind:options={field.options} />
                {:else if field.type === "date"}
                    <DateOptions {key} bind:options={field.options} />
                {:else if field.type === "select"}
                    <SelectOptions {key} bind:options={field.options} />
                {:else if field.type === "json"}
                    <JsonOptions {key} bind:options={field.options} />
                {:else if field.type === "file"}
                    <FileOptions {key} bind:options={field.options} />
                {:else if field.type === "relation"}
                    <RelationOptions {key} bind:options={field.options} />
                {:else if field.type === "user"}
                    <UserOptions {key} bind:options={field.options} />
                {/if}
            </div>

            <div class="col-4">
                <Field class="form-field form-field-toggle m-0" name="requried" let:uniqueId>
                    <input type="checkbox" id={uniqueId} bind:checked={field.required} />
                    <label for={uniqueId}>Required</label>
                </Field>
            </div>

            <div class="col-4">
                {#if field.type !== "file"}
                    <Field class="form-field form-field-toggle m-0" name="unique" let:uniqueId>
                        <input type="checkbox" id={uniqueId} bind:checked={field.unique} />
                        <label for={uniqueId}>Unique</label>
                    </Field>
                {/if}
            </div>
        </div>

        <input type="submit" class="hidden" tabindex="-1" />
    </form>
</Accordion>

<style>
    .title.field-name {
        max-width: 130px;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }
</style>
