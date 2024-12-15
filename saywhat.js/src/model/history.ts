/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import { isNode, SettingsStore } from './settings'
import { GeneratedItem, SpeechSettings } from './speech'

export class HistoryStore {
    static history: GeneratedItem[] = []
    static callbacks: (() => void)[] = []
    static subscribe(callback: () => void) {
        HistoryStore.callbacks = [...HistoryStore.callbacks, callback]
        const unsub = SettingsStore.subscribeKeyChange(() => HistoryStore.reloadMissingBlobs())
        return () => {
            HistoryStore.callbacks = HistoryStore.callbacks.filter((c) => c !== callback)
            unsub()
        }
    }
    static notify() {
        HistoryStore.callbacks.map((c) => c())
    }
    static getSnapshot() {
        if (HistoryStore.history.length == 0) {
            const loaded = HistoryStore.loadHistory()
            if (loaded.length > 0) {
                HistoryStore.history = loaded
                HistoryStore.reloadMissingBlobs()
            }
        }
        return HistoryStore.history
    }
    static saveHistory() {
        if (isNode) {
            console.log(JSON.stringify(HistoryStore.history, null, 4))
        } else {
            let withoutUrls = HistoryStore.history.map(
                (gi) => ({ ...gi, blob_url: '' }) as GeneratedItem,
            )
            localStorage.setItem('say_what_history', JSON.stringify(withoutUrls))
        }
    }
    static loadHistory() {
        if (!isNode) {
            const stored = localStorage.getItem('say_what_history')
            if (stored) {
                return JSON.parse(stored) as GeneratedItem[]
            }
        }
        return [] as GeneratedItem[]
    }
    static async loadMissingBlobs() {
        let missing = 0
        for (const item of HistoryStore.history) {
            if (item.blob_url) {
                continue
            }
            missing++
            const blob = await getHistoryItemAudio(item.history_item_id)
            item.blob_url = window.URL.createObjectURL(blob)
        }
        return missing > 0
    }
    static reloadMissingBlobs() {
        // don't need a reload if none are missing
        if (HistoryStore.history.filter((gi) => !gi.blob_url).length == 0) {
            return
        }
        // can't load blobs if we can't access the API
        if (SettingsStore.getSnapshot().settings.api_key.length < 32) {
            return
        }
        HistoryStore.loadMissingBlobs().then((ready) => ready && HistoryStore.updateAudio())
    }
    static addToHistory(gi: GeneratedItem) {
        HistoryStore.history = [gi, ...HistoryStore.history]
        HistoryStore.saveHistory()
        HistoryStore.notify()
    }
    static removeFromHistory(gi: GeneratedItem) {
        HistoryStore.history = HistoryStore.history.filter((hg) => hg !== gi)
        HistoryStore.saveHistory()
        HistoryStore.notify()
    }
    static updateFavorites() {
        HistoryStore.history = [...HistoryStore.history]
        HistoryStore.saveHistory()
        HistoryStore.notify()
    }
    static updateAudio() {
        HistoryStore.history = [...HistoryStore.history]
        HistoryStore.saveHistory()
        HistoryStore.notify()
    }
}

export interface HistoryItem {
    history_item_id: string
    request_id: string
    voice_id: string
    model_id: string
    voice_name: string
    voice_category: string
    text: string
    date_unix: number
    content_type: string
    state: string
    settings: SpeechSettings
}

interface HistoryPage {
    history: HistoryItem[]
    last_history_item_id: string
    has_more: boolean
}

export async function getHistoryItems(limit: number = 100) {
    const { settings } = SettingsStore.getSnapshot()
    const items: HistoryItem[] = []
    const endpoint = `${settings.api_root}/history`
    const method = 'GET'
    const headers = { 'xi-api-key': settings.api_key }
    let start_after_history_item_id: string = ''
    for (let needed = limit; needed > 0; ) {
        const page_size = (needed < 100 ? needed : 100).toString()
        const query: { [p: string]: string } = start_after_history_item_id
            ? { page_size, start_after_history_item_id }
            : { page_size }
        const url = endpoint + '?' + new URLSearchParams(query).toString()
        const response = await fetch(url, { method, headers })
        if (!response.ok) {
            const err = JSON.stringify(await response.json())
            let message = `${url} got (${response.status}): ${err}`
            console.error(message)
            throw Error(message)
        }
        const body: HistoryPage = await response.json()
        if (!body.has_more) {
            break
        }
        start_after_history_item_id = body.last_history_item_id
        const newItems = body.history
        items.push(...newItems)
        needed -= newItems.length
    }
    return items
}

export async function getHistoryItemAudio(id: string) {
    const { settings } = SettingsStore.getSnapshot()
    const url = `${settings.api_root}/history/${id}/audio`
    const method = 'GET'
    const headers = { 'xi-api-key': settings.api_key }
    const response = await fetch(url, { method, headers })
    if (!response.ok) {
        const err = JSON.stringify(await response.json())
        let message = `${url} got (${response.status}): ${err}`
        console.error(message)
        throw Error(message)
    }
    const mimeType = response.headers.get('Content-Type')
    if (!mimeType) {
        throw Error(`Audio mime type is: ${mimeType}`)
    }
    const blob = await response.blob()
    if (blob.type != mimeType) {
        console.warn(`Blob type (${blob.type}) isn't ${mimeType}`)
    }
    return blob
}
