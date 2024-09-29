import { HistoryStore } from './history'
import { ApiExternalStore } from './externalStore'
import { SettingsStore } from './settings'

export interface Voice {
    voice_id: string
    name: string
}

interface VoicePage {
    voices: Voice[]
}

export async function getVoices() {
    const { settings } = SettingsStore.getSnapshot()
    const url = `${settings.api_root}/voices`
    const method = 'GET'
    const headers = { 'xi-api-key': settings.api_key }
    const response = await fetch(url, { method, headers })
    if (!response.ok) {
        let detail = JSON.stringify(await response.json())
        let message = `${url} got ${response.status}: ${detail}`
        console.error(message)
        return []
    }
    const page: VoicePage = await response.json()
    return page.voices
}

export interface PdictMetadata {
    id: string
    latest_version_id: string
    name: string
    creation_time_unix: string
    description: string
}

function metadataToOption(md: PdictMetadata): PdictOption {
    return {
        id: `${md.id}|${md.latest_version_id}`,
        name: md?.description ? md.description : md.name,
    }
}

export interface PdictOption {
    id: string
    name: string
}

export interface PdictLocator {
    pronunciation_dictionary_id: string
    version_id: string
}

function stringToLocator(s: string) {
    let parts = s.split('|')
    if (parts.length != 2) {
        return undefined
    }
    const value: PdictLocator = {
        pronunciation_dictionary_id: parts[0],
        version_id: parts[1],
    }
    return value
}

interface PdictMetadataPage {
    pronunciation_dictionaries: PdictMetadata[]
}

export async function getPdicts() {
    const fallback: PdictOption = { id: '', name: '(None)' }
    const { settings } = SettingsStore.getSnapshot()
    const url = `${settings.api_root}/pronunciation-dictionaries/`
    const method = 'GET'
    const headers = { 'xi-api-key': settings.api_key }
    const response = await fetch(url, { method, headers })
    if (!response.ok) {
        let detail = JSON.stringify(await response.json())
        let message = `${url} got ${response.status}: ${detail}`
        console.error(message)
        return [fallback]
    }
    const page: PdictMetadataPage = await response.json()
    const dicts: PdictOption[] = page.pronunciation_dictionaries.map(metadataToOption)
    return [fallback, ...dicts]
}

export interface AutoCompleteOption {
    label: string
    id: string
}

export const voiceStore = new ApiExternalStore<AutoCompleteOption>(async () => {
    let voices = await getVoices()
    let options = voices.map(
        (voice) => ({ label: voice.name, id: voice.voice_id }) as AutoCompleteOption,
    )
    options.sort((a, b) => (a.label < b.label ? -1 : b.label < a.label ? 1 : 0))
    return options
})

export const pdictStore = new ApiExternalStore<AutoCompleteOption>(async () => {
    let pdicts = await getPdicts()
    let options = pdicts.map((pdict) => ({ label: pdict.name, id: pdict.id }) as AutoCompleteOption)
    options.sort((a, b) => (a.label < b.label ? -1 : b.label < a.label ? 1 : 0))
    return options
})

export interface Language {
    language_id: string
    name: string
}

export interface Model {
    can_be_finetuned: boolean
    can_do_text_to_speech: boolean
    can_do_voice_conversion: boolean
    can_use_speaker_boost: boolean
    can_use_style: boolean
    description: string
    languages: Language[]
    max_characters_request_free_user: number
    max_characters_request_subscribed_user: number
    model_id: string
    name: string
    requires_alpha_access: boolean
    serves_pro_voices: boolean
    token_cost_factor: number
}

export async function getModels() {
    const { settings } = SettingsStore.getSnapshot()
    const url = `${settings.api_root}/models`
    const method = 'GET'
    const headers = { 'xi-api-key': settings.api_key }
    const response = await fetch(url, { method, headers })
    if (!response.ok) {
        let detail = JSON.stringify(await response.json())
        let message = `${url} got ${response.status}: ${detail}`
        console.error(message)
        return []
    }
    const models: Model[] = await response.json()
    return models
}

export const modelStore = new ApiExternalStore<AutoCompleteOption>(async () => {
    let models = await getModels()
    let options = models.map(
        (model) => ({ label: model.name, id: model.model_id }) as AutoCompleteOption,
    )
    options.sort((a, b) => (a.label < b.label ? -1 : b.label < a.label ? 1 : 0))
    return options
})

export interface SpeechSettings {
    similarity_boost: number
    stability: number
    use_speaker_boost: boolean
}

export interface GenerationSettings {
    output_format: string
    optimize_streaming_latency: string
    voice_id: string
    model_id: string
    voice_settings: SpeechSettings
    pronunciation_dictionary: string
}

export function generationSettingsEqual(s1: GenerationSettings, s2: GenerationSettings) {
    if (s1 === s2) return true
    if (s1.output_format != s2.output_format) return false
    if (s1.optimize_streaming_latency != s2.optimize_streaming_latency) return false
    if (s1.voice_id != s2.voice_id) return false
    if (s1.model_id != s2.model_id) return false
    if (s1.voice_settings === s2.voice_settings) return true
    if (s1.voice_settings.stability != s2.voice_settings.stability) return false
    if (s1.voice_settings.similarity_boost != s2.voice_settings.similarity_boost) return false
    if (s1.voice_settings.use_speaker_boost != s2.voice_settings.use_speaker_boost) return false
    if (s1.pronunciation_dictionary != s2.pronunciation_dictionary) return false
    return true
}

export interface GeneratedItem {
    history_item_id: string
    text: string
    settings: GenerationSettings
    gen_ms: number
    gen_date: number
    kb_blob_size: number
    blob_url: string
    favorite: boolean
}

export async function generateSpeech(text: string) {
    const { settings } = SettingsStore.getSnapshot()
    const voiceId = settings.generation_settings.voice_id
    const endpoint = `${settings.api_root}/text-to-speech/${voiceId}`
    const method = 'POST'
    const headers = {
        'xi-api-key': settings.api_key,
        'Content-Type': 'application/json',
    }
    const query: { [p: string]: string } = {
        output_format: settings.generation_settings.output_format,
        optimize_streaming_latency: settings.generation_settings.optimize_streaming_latency,
    }
    const url = endpoint + '?' + new URLSearchParams(query).toString()
    const params: { [k: string]: any } = {
        model_id: settings.generation_settings.model_id,
        text,
        voice_settings: {
            similarity_boost: settings.generation_settings.voice_settings.similarity_boost,
            stability: settings.generation_settings.voice_settings.stability,
            use_speaker_boost: settings.generation_settings.voice_settings.use_speaker_boost,
        },
    }
    const dictParam = stringToLocator(settings.generation_settings.pronunciation_dictionary)
    if (dictParam) {
        params.pronunciation_dictionary_locators = [dictParam]
    }
    const body = JSON.stringify(params)
    const start = performance.now()
    const response = await fetch(url, { method, headers, body })
    if (!response.ok) {
        const err = JSON.stringify(await response.json())
        let message = `${url} got (${response.status}): ${err}`
        console.error(message)
        throw Error(message)
    }
    const id = response.headers.get('history-item-id')
    const blob = await response.blob()
    const elapsed = performance.now() - start
    const audio = URL.createObjectURL(blob)
    const item: GeneratedItem = {
        history_item_id: id || '',
        text,
        settings: settings.generation_settings,
        gen_ms: Math.ceil(elapsed),
        gen_date: Date.now().valueOf(),
        kb_blob_size: Math.ceil(blob.size / 1024),
        blob_url: audio,
        favorite: false,
    }
    HistoryStore.addToHistory(item)
    return item
}
