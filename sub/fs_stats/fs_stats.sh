#!/bin/bash
iostat_fsStats_info() {
    output_file="fs_stats.stats"
    export S_TIME_FORMAT=ISO
    fs_stat_info=$(iostat -d --human | tail -n +4)
    echo "fsStats:" > "$output_file"
    while read -r name tps rps wps read_kb write_kb; do
        if [[ $name =~ ^[hsv]d[a-z]$ ]]; then
            cat >> $output_file <<EOL
    - device: ${name}
      rxPerSec: ${rps}
      wxPerSec: ${wps}
      source: iostat
EOL
        fi
    done <<< "$fs_stat_info"
    unset S_TIME_FORMAT
}
gua_fsStats_info() {
    output_file="fs_stats.stats"
    tmp_data_file=~/.xterminal/fs_stats.tmp
    local disk_name_regex='^(sd[a-z]+|vd[a-z]+|hd[a-z]+|xvd[a-z]+|nvme[0-9]+n[0-9]+)$'
    local prev_time=''
    local prev_stats=''
    if [[ -f $tmp_data_file ]]; then
        local first_line=''
        first_line=$(head -n 1 "$tmp_data_file" 2>/dev/null)
        local first_line_stats=''
        if [[ -n $first_line ]]; then
            read -r prev_time first_line_stats <<< "$first_line"
        fi
        prev_stats=$(tail -n +2 "$tmp_data_file" 2>/dev/null)
        if [[ -n $first_line_stats ]]; then
            if [[ -n $prev_stats ]]; then
                prev_stats="${first_line_stats}
${prev_stats}"
            else
                prev_stats="$first_line_stats"
            fi
        fi
        if [[ ! $prev_time =~ ^[0-9]+$ ]]; then
            prev_time=''
            prev_stats=''
        fi
    fi
    local curr_time
    curr_time=$(date +%s)
    local curr_stats
    curr_stats=$(awk -v re="$disk_name_regex" '$3 ~ re {print $3,$6,$10}' /proc/diskstats 2>/dev/null)
    {
        echo "$curr_time"
        echo "$curr_stats"
    } > "$tmp_data_file"
    echo "fsStats:" > "$output_file"
    if [[ -z $prev_time || -z $prev_stats ]]; then
        while read -r dev_name _ _; do
            [[ -n $dev_name ]] || continue
            echo "  - device: $dev_name" >> "$output_file"
            echo "    rxPerSec: 0" >> "$output_file"
            echo "    wxPerSec: 0" >> "$output_file"
            echo "    source: gua" >> "$output_file"
        done <<< "$curr_stats"
        return
    fi
    local interval=$((curr_time - prev_time))
    if [[ $interval -le 0 ]]; then
        interval=1
    fi
    declare -A prev_reads_by_dev
    declare -A prev_writes_by_dev
    while read -r dev_name sectors_read sectors_write; do
        [[ -n $dev_name ]] || continue
        prev_reads_by_dev["$dev_name"]=$sectors_read
        prev_writes_by_dev["$dev_name"]=$sectors_write
    done <<< "$prev_stats"
    while read -r dev_name curr_reads curr_writes; do
        [[ -n $dev_name ]] || continue
        local rxPerSec=0
        local wxPerSec=0
        local prev_reads=${prev_reads_by_dev["$dev_name"]}
        local prev_writes=${prev_writes_by_dev["$dev_name"]}
        if [[ -n $prev_reads && -n $prev_writes ]]; then
            local sector_size=512
            if [[ -r /sys/block/$dev_name/queue/hw_sector_size ]]; then
                sector_size=$(cat /sys/block/$dev_name/queue/hw_sector_size 2>/dev/null)
            fi
            rxPerSec=$(( ((curr_reads - prev_reads) * sector_size) / interval ))
            wxPerSec=$(( ((curr_writes - prev_writes) * sector_size) / interval ))
            if [[ $rxPerSec -lt 0 ]]; then rxPerSec=0; fi
            if [[ $wxPerSec -lt 0 ]]; then wxPerSec=0; fi
        fi
        echo "  - device: $dev_name" >> "$output_file"
        echo "    rxPerSec: $rxPerSec" >> "$output_file"
        echo "    wxPerSec: $wxPerSec" >> "$output_file"
        echo "    source: gua" >> "$output_file"
    done <<< "$curr_stats"
}
__main() {
    gua_fsStats_info
}
__main
