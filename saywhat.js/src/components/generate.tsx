/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import React, { useState } from 'react'
import { generateSpeech } from '../model/speech'
import Typography from '@mui/material/Typography'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import { Settings } from '../model/settings'

export function Generate(props: { settings: Settings }) {
    const [text, setText] = useState('')
    const [state, setState] = useState('Please enter text')
    const [error, setError] = useState('')

    function generate() {
        setState('Processing...')
        setError('')
        console.log(`Generating voice for: ${text}`)
        generateSpeech(text)
            .then((gi) => {
                console.log(`Generation complete in ${gi.gen_ms} milliseconds.`)
                setState('Ready to generate')
            })
            .catch((error) => {
                setError(error.toString())
                setState(`Generation failure: Click to try again`)
            })
    }

    function onChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
        let val = e.target.value
        setText(val)
        setState(val ? 'Ready to generate' : 'Enter text')
        setError('')
    }

    return (
        <>
            {props.settings.api_key ? (
                <>
                    <TextField multiline fullWidth label="Text to Speak" onChange={onChange} />
                    <Button variant="contained" onClick={generate} disabled={!text}>
                        {state}
                    </Button>
                    {error && <Typography>{error}</Typography>}
                </>
            ) : (
                <Typography>An API key is required for generation</Typography>
            )}
        </>
    )
}
