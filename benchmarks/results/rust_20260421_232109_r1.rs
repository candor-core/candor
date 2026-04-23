```rust
use std::collections::HashMap;
use std::env;
use std::fs;
use std::io::{self, BufRead};
use std::process;

#[derive(Debug, Default)]
struct FileStats {
    error_count: u64,
    warn_count: u64,
    info_count: u64,
    debug_count: u64,
    parse_errors: u64,
}

enum ParsedLine {
    Error,
    Warn,
    Info,
    Debug,
    ParseError,
}

fn parse_line(line: &str) -> ParsedLine {
    let trimmed = line.trim();
    if trimmed.is_empty() {
        return ParsedLine::ParseError;
    }

    if !trimmed.starts_with('[') {
        return ParsedLine::ParseError;
    }

    let close_bracket = match trimmed.find(']') {
        Some(pos) => pos,
        None => return ParsedLine::ParseError,
    };

    let level = &trimmed[1..close_bracket];

    match level {
        "ERROR" => ParsedLine::Error,
        "WARN" => ParsedLine::Warn,
        "INFO" => ParsedLine::Info,
        "DEBUG" => ParsedLine::Debug,
        _ => ParsedLine::ParseError,
    }
}

fn aggregate_file_stats(path: &std::path::Path) -> io::Result<FileStats> {
    let file = fs::File::open(path)?;
    let reader = io::BufReader::new(file);
    let mut stats = FileStats::default();

    for line_result in reader.lines() {
        let line = line_result?;
        if line.trim().is_empty() {
            continue;
        }
        match parse_line(&line) {
            ParsedLine::Error => stats.error_count += 1,
            ParsedLine::Warn => stats.warn_count += 1,
            ParsedLine::Info => stats.info_count += 1,
            ParsedLine::Debug => stats.debug_count += 1,
            ParsedLine::ParseError => stats.parse_errors += 1,
        }
    }

    Ok(stats)
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        eprintln!("Usage: {} <input_directory>", args[0]);
        process::exit(1);
    }

    let dir_path = &args[1];

    let entries = match fs::read_dir(dir_path) {
        Ok(e) => e,
        Err(err) => {
            eprintln!("Error reading directory '{}': {}", dir_path, err);
            process::exit(1);
        }
    };

    let mut log_files: Vec<String> = Vec::new();

    for entry_result in entries {
        match entry_result {
            Ok(entry) => {
                let file_name = entry.file_name();
                let file_name_str = file_name.to_string_lossy().to_string();
                if file_name_str.ends_with(".log") {
                    let metadata = match entry.metadata() {
                        Ok(m) => m,
                        Err(_) => continue,
                    };
                    if metadata.is_file() {
                        log_files.push(file_name_str);
                    }
                }
            }
            Err(err) => {
                eprintln!("Error reading directory entry: {}", err);
            }
        }
    }

    log_files.sort();

    let mut file_stats_map: HashMap<String, FileStats> = HashMap::new();

    for file_name in &log_files {
        let mut full_path = std::path::PathBuf::from(dir_path);
        full_path.push(file_name);

        match aggregate_file_stats(&full_path) {
            Ok(stats) => {
                file_stats_map.insert(file_name.clone(), stats);
            }
            Err(err) => {
                eprintln!("Error processing file '{}': {}", file_name, err);
                file_stats_map.insert(file_name.clone(), FileStats::default());
            }
        }
    }

    let mut total_files: u64 = 0;
    let mut total_lines: u64 = 0;
    let mut total_parse_errors: u64 = 0;

    for file_name in &log_files {
        let stats = &file_stats_map[file_name];
        println!(
            "{}: ERROR={} WARN={} INFO={} DEBUG={} parse_errors={}",
            file_name,
            stats.error_count,
            stats.warn_count,
            stats.info_count,
            stats.debug_count,
            stats.parse_errors
        );
        total_files += 1;
        total_lines += stats.error_count
            + stats.warn_count
            + stats.info_count
            + stats.debug_count
            + stats.parse_errors;
        total_parse_errors += stats.parse_errors;
    }

    println!(
        "TOTAL: files={} lines={} parse_errors={}",
        total_files, total_lines, total_parse_errors
    );

    process::exit(0);
}
```