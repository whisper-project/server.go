import * as React from 'react'

import TextField from '@mui/material/TextField'
import Autocomplete from '@mui/material/Autocomplete'

interface Option {
    label: string
    id: string
}

export function Picker(props: {
    name: string
    options: Option[]
    initial: string
    label: string
    updater: (val: string) => void
}) {
    const onChange = (e: React.SyntheticEvent, v: Option) => {
        props.updater(v.id)
    }
    return (
        <Autocomplete
            fullWidth={true}
            id={props.name}
            options={props.options}
            value={props.options.find((o) => o.id == props.initial) || props.options[0]}
            onChange={onChange}
            disableClearable={true}
            renderInput={(params) => <TextField {...params} label={props.label} />}
        />
    )
}
