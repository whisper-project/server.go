import * as React from 'react'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import Box from '@mui/material/Box'
import Grid from '@mui/material/Grid'

export function ToggleSwitch(props: {
    name: string
    initial: boolean
    label: string
    updater: (val: boolean) => void
}) {
    const [checked, setChecked] = React.useState(props.initial)

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        props.updater(event.target.checked)
        setChecked(event.target.checked)
    }

    return (
        <Box sx={{ width: 250 }}>
            <Grid container alignItems="left">
                <FormControlLabel
                    labelPlacement="start"
                    control={
                        <Switch
                            checked={checked}
                            onChange={handleChange}
                            inputProps={{ 'aria-label': `switch-${props.name}` }}
                        />
                    }
                    label={props.label}
                />
            </Grid>
        </Box>
    )
}
