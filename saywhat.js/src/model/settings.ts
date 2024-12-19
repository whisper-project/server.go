/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import fnv from 'fnv-plus'

import { GenerationSettings, generationSettingsEqual } from './speech'

const whisperApiRoot = '/api/say-what/v1'

export const isNode =
    typeof process !== 'undefined' && typeof process?.versions?.node !== 'undefined'

export interface Profile {
    id: string
    password: string
    serverPassword: string
    serverETag: string
}

export interface Settings {
    api_key: string
    api_root: string
    generation_settings: GenerationSettings
}

export interface CachedSettings {
    profile: Profile
    settings: Settings
}

function settingsETag(settings: Settings) {
    let string = settings.api_key
    string += '|' + settings.api_root
    string += '|' + settings.generation_settings.output_format
    string += '|' + settings.generation_settings.optimize_streaming_latency
    string += '|' + settings.generation_settings.voice_id
    string += '|' + settings.generation_settings.model_id
    string += '|' + settings.generation_settings.voice_settings.similarity_boost.toString()
    string += '|' + settings.generation_settings.voice_settings.stability.toString()
    string += '|' + settings.generation_settings.pronunciation_dictionary
    return fnv.fast1a32hex(string)
}

const defaultSettings: Settings = {
    api_key: '',
    api_root: 'https://api.elevenlabs.io/v1',
    generation_settings: {
        output_format: 'mp3_44100_128',
        optimize_streaming_latency: '0',
        voice_id: 'pNInz6obpgDQGcFmaJgB', // Adam
        model_id: 'eleven_turbo_v2',
        voice_settings: {
            similarity_boost: 0.5,
            stability: 0.5,
            use_speaker_boost: true,
        },
        pronunciation_dictionary: '',
    },
}

const defaultProfile: Profile = {
    id: '',
    password: '',
    serverPassword: '',
    serverETag: '',
}

export class SettingsStore {
    static profile: Profile
    static settings: Settings
    static cached: CachedSettings
    static eTag: string
    static profileChangeCallbacks: (() => void)[] = []
    static keyChangeCallbacks: (() => void)[] = []
    static anyChangeCallbacks: (() => void)[] = []

    static subscribeProfileChange(cfn: () => void) {
        SettingsStore.profileChangeCallbacks = [...SettingsStore.profileChangeCallbacks, cfn]
        return () => {
            SettingsStore.profileChangeCallbacks = SettingsStore.profileChangeCallbacks.filter(
                (c) => c !== cfn,
            )
        }
    }

    static subscribeKeyChange(cfn: () => void) {
        SettingsStore.keyChangeCallbacks = [...SettingsStore.keyChangeCallbacks, cfn]
        return () => {
            SettingsStore.keyChangeCallbacks = SettingsStore.keyChangeCallbacks.filter(
                (c) => c !== cfn,
            )
        }
    }

    static subscribe(cfn: () => void) {
        SettingsStore.anyChangeCallbacks = [...SettingsStore.anyChangeCallbacks, cfn]
        return () => {
            SettingsStore.anyChangeCallbacks = SettingsStore.anyChangeCallbacks.filter(
                (c) => c !== cfn,
            )
        }
    }

    static notifyProfileChange() {
        SettingsStore.profileChangeCallbacks.map((c) => c())
    }

    static notifyKeyChange() {
        SettingsStore.keyChangeCallbacks.map((c) => c())
    }
    static notifyChange() {
        SettingsStore.anyChangeCallbacks.map((c) => c())
    }

    static updateLocalProfile(profileId: string, profilePassword: string) {
        const { id, password } = SettingsStore.profile
        if (profileId === id && profilePassword === password) {
            return
        }
        SettingsStore.profile = {
            id: profileId,
            password: profilePassword,
            serverPassword: '',
            serverETag: '',
        }
        SettingsStore.saveProfile()
        SettingsStore.cached = { profile: SettingsStore.profile, settings: SettingsStore.settings }
        SettingsStore.notifyChange()
        SettingsStore.downloadProfile().then()
    }

    static updateLocalSettings(
        apiKey: string,
        outputFormat: string,
        optimizeStreamingLatency: string,
        voiceId: string,
        modelId: string,
        similarityBoost: number,
        stability: number,
        useSpeakerBoost: boolean,
        pdictId: string,
    ) {
        const generationSettings = {
            output_format: outputFormat,
            optimize_streaming_latency: optimizeStreamingLatency,
            voice_id: voiceId,
            model_id: modelId,
            voice_settings: {
                similarity_boost: similarityBoost,
                stability: stability,
                use_speaker_boost: useSpeakerBoost,
            },
            pronunciation_dictionary: pdictId,
        }
        const newSettings = {
            api_key: apiKey,
            api_root: SettingsStore.settings.api_root,
            generation_settings: generationSettings,
        }
        if (SettingsStore.updateSettings(newSettings)) {
            SettingsStore.uploadProfile().then()
        }
    }

    static updateLocalGenerationSettings(settings: GenerationSettings) {
        const newSettings = {
            api_key: SettingsStore.settings.api_key,
            api_root: SettingsStore.settings.api_root,
            generation_settings: settings,
        }
        if (SettingsStore.updateSettings(newSettings)) {
            SettingsStore.uploadProfile().then()
        }
    }

    static updateSettings(data: Settings) {
        const notifyApiKeyChange = data.api_key != SettingsStore.settings.api_key
        const notifyGenerationChange = !generationSettingsEqual(
            SettingsStore.settings.generation_settings,
            data.generation_settings,
        )
        if (!notifyApiKeyChange || notifyGenerationChange) {
            return false
        }
        SettingsStore.eTag = settingsETag(data)
        SettingsStore.settings = data
        SettingsStore.saveSettings()
        SettingsStore.cached = { profile: SettingsStore.profile, settings: SettingsStore.settings }
        if (notifyApiKeyChange) {
            SettingsStore.notifyKeyChange()
        }
        SettingsStore.notifyChange()
        return true
    }

    static getSnapshot() {
        if (!SettingsStore.cached) {
            SettingsStore.loadSettings()
            SettingsStore.loadProfile()
            SettingsStore.cached = {
                profile: SettingsStore.profile,
                settings: SettingsStore.settings,
            }
        }
        return SettingsStore.cached
    }

    static loadProfile() {
        SettingsStore.profile = defaultProfile
        if (!isNode) {
            const stored = localStorage.getItem('say_what_profileInfo')
            if (stored !== null && stored.length > 0) {
                SettingsStore.profile = JSON.parse(stored)
                SettingsStore.downloadProfile().then()
            }
        }
    }

    static saveProfile() {
        if (!isNode) {
            localStorage.setItem('say_what_profileInfo', JSON.stringify(SettingsStore.profile))
        }
    }

    static loadSettings() {
        SettingsStore.settings = defaultSettings
        SettingsStore.eTag = settingsETag(defaultSettings)
        if (!isNode) {
            const stored = localStorage.getItem('say_what_settings')
            if (stored !== null && stored.length > 0) {
                const restored: Settings = JSON.parse(stored)
                SettingsStore.settings = restored
                SettingsStore.eTag = settingsETag(restored)
                return
            }
        }
    }

    static saveSettings() {
        if (!isNode) {
            localStorage.setItem('say_what_settings', JSON.stringify(SettingsStore.settings))
        }
    }

    static async downloadProfile() {
        let { id, password, serverPassword, serverETag } = SettingsStore.profile
        if (!id || !password) {
            return
        }
        if (!serverPassword) {
            const sha1 = await crypto.subtle.digest('SHA-1', new TextEncoder().encode(password))
            serverPassword = Buffer.from(sha1).toString('hex')
        }
        const headers = new Headers()
        headers.append('Authorization', `Bearer ${serverPassword}`)
        if (serverETag) {
            headers.append('If-None-Match', `"${serverETag}"`)
        }
        const resp = await fetch(whisperApiRoot + `/settings/${id}`, {
            method: 'GET',
            mode: 'cors',
            headers,
        }).catch((err) => {
            console.error(`Network error on profile GET: ${err}`)
            return new Response('', {
                status: 500,
                statusText: `Network error reaching ${whisperApiRoot}`,
            })
        })
        if (resp.status == 404) {
            SettingsStore.profile.serverETag = ''
            await SettingsStore.uploadProfile()
            return
        }
        if (resp.status == 403) {
            console.error(`Incorrect password on profile download`)
            SettingsStore.profile = { id, password: '', serverPassword: '', serverETag: '' }
            SettingsStore.notifyChange()
            return
        }
        if (resp.status == 304 || resp.status == 412) {
            // settings are up to date
            return
        }
        if (resp.status != 200) {
            console.error(`Received unexpected status ${resp.status} on profile GET`)
            return
        }
        const data: Settings = await resp.json()
        SettingsStore.profile.serverETag = settingsETag(data)
        SettingsStore.updateSettings(data)
    }

    static async uploadProfile() {
        const { id, serverPassword, serverETag } = SettingsStore.profile
        if (!id || !serverPassword) {
            return
        }
        const method = serverETag ? 'PUT' : 'POST'
        const headers = new Headers()
        headers.append('Content-Type', 'application/json')
        if (method === 'PUT') {
            headers.append('Authorization', `Bearer ${serverPassword}`)
            headers.append('If-None-Match', `"${SettingsStore.eTag}"`)
        }
        const resp = await fetch(whisperApiRoot + `/settings/${id}`, {
            method,
            mode: 'same-origin',
            headers,
            body: JSON.stringify(SettingsStore.settings),
        }).catch((err) => {
            console.error(`Network error on profile GET: ${err}`)
            return new Response('', {
                status: 500,
                statusText: `Network error reaching ${whisperApiRoot}`,
            })
        })
        if (resp.status == 403) {
            console.error(`Incorrect password on profile upload`)
            SettingsStore.profile = { id, password: '', serverPassword: '', serverETag: '' }
            SettingsStore.notifyChange()
            return
        }
        if (resp.status == 304 || resp.status == 412) {
            // no change on internal side
            SettingsStore.profile.serverETag = SettingsStore.eTag
            return
        }
        if (resp.status != 201 && resp.status != 204) {
            console.error(`Received unexpected status ${resp.status} on profile ${method}`)
            return
        }
        const eTag = resp.headers.get('ETag')
        if (eTag === null || eTag.length < 3) {
            console.warn(`Received no eTag on profile upload`)
            SettingsStore.profile.serverETag = SettingsStore.eTag
        } else {
            // remove the quotes
            SettingsStore.profile.serverETag = eTag.substring(1, eTag.length - 1)
        }
    }
}
