/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import React, { useEffect, useState, useSyncExternalStore } from 'react'

import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'
import Grid from '@mui/material/Grid'

import { Settings, SettingsStore } from '../model/settings'
import { modelStore, pdictStore, voiceStore } from '../model/speech'

import { ZeroToOneSlider } from './slider'
import { Picker } from './picker'

export function SettingsView(props: { settings: Settings }) {
    const [apiKey, setApiKey] = useState(props.settings.api_key)
    const [format, setFormat] = useState(props.settings.generation_settings.output_format)
    const [latency, setLatency] = useState(
        props.settings.generation_settings.optimize_streaming_latency,
    )
    const [voiceId, setVoiceId] = useState(props.settings.generation_settings.voice_id)
    const [modelId, setModelId] = useState(props.settings.generation_settings.model_id)
    const [similarity, setSimilarity] = useState(
        props.settings.generation_settings.voice_settings.similarity_boost,
    )
    const [stability, setStability] = useState(
        props.settings.generation_settings.voice_settings.stability,
    )
    const [boost, setBoost] = useState(
        props.settings.generation_settings.voice_settings.use_speaker_boost,
    )
    const [pdictId, setPdictId] = useState(
        props.settings.generation_settings.pronunciation_dictionary,
    )
    useEffect(() => {
        SettingsStore.updateLocalSettings(
            apiKey,
            format,
            latency,
            voiceId,
            modelId,
            similarity,
            stability,
            boost,
            pdictId,
        )
    }, [apiKey, format, latency, voiceId, modelId, similarity, stability, boost, pdictId])
    return (
        <>
            {apiKey && (
                <>
                    <VoiceModelSettings
                        voiceId={voiceId}
                        setVoiceId={setVoiceId}
                        modelId={modelId}
                        setModelId={setModelId}
                    />
                    <FormatLatencySettings
                        format={format}
                        setFormat={setFormat}
                        latency={latency}
                        setLatency={setLatency}
                    />
                    <VoiceSettings
                        stability={stability}
                        setStability={setStability}
                        similarity={similarity}
                        setSimilarity={setSimilarity}
                        boost={boost}
                        setBoost={setBoost}
                    />
                    <PronunciationSettings pdictId={pdictId} setPdictId={setPdictId} />
                </>
            )}
            <ApiKey apiKey={apiKey} setApiKey={setApiKey} />
        </>
    )
}

function ApiKey(props: {
    apiKey: string
    setApiKey: React.Dispatch<React.SetStateAction<string>>
}) {
    const [input, setInput] = useState(props.apiKey)
    const onSubmit = () => props.setApiKey(input)
    const onChange = (e: React.ChangeEvent<HTMLInputElement>) => setInput(e.target.value)
    return (
        <Grid container component="form" noValidate autoComplete="off">
            <Grid item>
                <TextField
                    id="outlined-basic"
                    label="ElevenLabs API Key"
                    variant="outlined"
                    style={{ width: '54ch' }}
                    value={input}
                    onChange={onChange}
                />
            </Grid>
            <Grid item alignItems="stretch" style={{ display: 'flex' }}>
                <Button variant="contained" onClick={onSubmit}>
                    Set API Key
                </Button>
            </Grid>
        </Grid>
    )
}

function VoiceModelSettings(props: {
    voiceId: string
    setVoiceId: React.Dispatch<React.SetStateAction<string>>
    modelId: string
    setModelId: React.Dispatch<React.SetStateAction<string>>
}) {
    const voices = useSyncExternalStore(
        (c) => voiceStore.subscribe(c),
        () => voiceStore.getSnapshot(),
    )
    const models = useSyncExternalStore(
        (c) => modelStore.subscribe(c),
        () => modelStore.getSnapshot(),
    )
    return (
        <>
            {voices.length ? (
                <Picker
                    name="VoiceId"
                    options={voices}
                    initial={props.voiceId}
                    label={'Voice'}
                    updater={props.setVoiceId}
                />
            ) : (
                <Typography>Retrieving voices...</Typography>
            )}
            {models.length ? (
                <Picker
                    name="ModelId"
                    options={models}
                    initial={props.modelId}
                    label={'Model'}
                    updater={props.setModelId}
                />
            ) : (
                <Typography>Retrieving models...</Typography>
            )}
        </>
    )
}

function FormatLatencySettings(props: {
    format: string
    setFormat: React.Dispatch<React.SetStateAction<string>>
    latency: string
    setLatency: React.Dispatch<React.SetStateAction<string>>
}) {
    // const formatOptions = [
    //     { label: '22.05kHz sample rate at 32kbps', id: 'mp3_22050_32' },
    //     { label: '44.1kHz sample rate at 32kbps', id: 'mp3_44100_32' },
    //     { label: '44.1kHz sample rate at 64kbps', id: 'mp3_44100_64' },
    //     { label: '44.1kHz sample rate at 96kbps', id: 'mp3_44100_96' },
    //     { label: '44.1kHz sample rate at 128kbps', id: 'mp3_44100_128' },
    // ]
    const latencyOptions = [
        { label: 'No latency optimization', id: '0' },
        { label: 'Normal (50% of max) latency optimization', id: '1' },
        { label: 'Strong (75% of max) latency optimization', id: '2' },
        { label: 'Max latency optimization', id: '3' },
    ]
    return (
        <>
            {/*<Picker*/}
            {/*    name="OutputFormat"*/}
            {/*    options={formatOptions}*/}
            {/*    initial={props.format}*/}
            {/*    label={'Output Format'}*/}
            {/*    updater={props.setFormat}*/}
            {/*/>*/}
            <Picker
                name="Latency"
                options={latencyOptions}
                initial={props.latency}
                label={'Latency'}
                updater={props.setLatency}
            />
        </>
    )
}

function VoiceSettings(props: {
    stability: number
    setStability: React.Dispatch<React.SetStateAction<number>>
    similarity: number
    setSimilarity: React.Dispatch<React.SetStateAction<number>>
    boost: boolean
    setBoost: React.Dispatch<React.SetStateAction<boolean>>
}) {
    return (
        <>
            <Grid container spacing={1} alignItems="left">
                <Grid item xs={6}>
                    <ZeroToOneSlider
                        name="Stability"
                        initial={props.stability}
                        label={'Stability'}
                        updater={props.setStability}
                    />
                </Grid>
                <Grid item xs={6}>
                    <ZeroToOneSlider
                        name="SimilarityBoost"
                        initial={props.similarity}
                        label={'Similarity Boost'}
                        updater={props.setSimilarity}
                    />
                </Grid>
                {/*<Grid item>*/}
                {/*    <ToggleSwitch*/}
                {/*        name={'speakerBoost'}*/}
                {/*        initial={props.boost}*/}
                {/*        label={'Speaker Boost'}*/}
                {/*        updater={props.setBoost}*/}
                {/*    />*/}
                {/*</Grid>*/}
            </Grid>
        </>
    )
}

function PronunciationSettings(props: {
    pdictId: string
    setPdictId: React.Dispatch<React.SetStateAction<string>>
}) {
    const dicts = useSyncExternalStore(
        (c) => pdictStore.subscribe(c),
        () => pdictStore.getSnapshot(),
    )
    return (
        <>
            {dicts.length ? (
                <Picker
                    name="DictionaryId"
                    options={dicts}
                    initial={props.pdictId}
                    label={'Pronunciation Dictionary'}
                    updater={props.setPdictId}
                />
            ) : (
                <Typography>Retrieving pronunciation dictionaries...</Typography>
            )}
        </>
    )
}
