use std::collections::HashMap;
use std::env;
use std::fs;
use std::path::Path;
use std::process;

#[derive(Debug, Default)]
struct FileStats {
    error: u64,
    warn: u64,
    info: u64,
    debug: u64,
    parse_errors: u64,
}

impl FileStats {
    fn total_lines(&self) -> u64 {
        self.error + self.warn + self.info + self.debug + self.parse_errors
    }
}

#[derive(Debug, PartialEq)]
enum LogLevel {
    Error,
    Warn,
    Info,
    Debug,
}

#[derive(Debug)]
enum ParseResult {
    Ok(LogLevel, String),
    Err,
}

fn parse_line(line: &str) -> ParseResult {
    let line = line.trim();
    if line.is_empty() {
        return ParseResult::Err;
    }

    if !line.starts_with('[') {
        return ParseResult::Err;
    }

    let close_bracket = match line.find(']') {
        Some(pos) => pos,
        None => return ParseResult::Err,
    };

    let level_str = &line[1..close_bracket];
    let rest = &line[close_bracket + 1..];

    if !rest.starts_with(' ') {
        return ParseResult::Err;
    }

    let message = &rest[1..];

    let level = match level_str {
        "ERROR" => LogLevel::Error,
        "WARN" => LogLevel::Warn,
        "INFO" => LogLevel::Info,
        "DEBUG" => LogLevel::Debug,
        _ => return ParseResult::Err,
    };

    ParseResult::Ok(level, message.to_string())
}

fn aggregate_stats(lines: &[&str]) -> FileStats {
    let mut stats = FileStats::default();

    for line in lines {
        if line.trim().is_empty() {
            continue;
        }

        match parse_line(line) {
            ParseResult::Ok(LogLevel::Error, _) => stats.error += 1,
            ParseResult::Ok(LogLevel::Warn, _) => stats.warn += 1,
            ParseResult::Ok(LogLevel::Info, _) => stats.info += 1,
            ParseResult::Ok(LogLevel::Debug, _) => stats.debug += 1,
            ParseResult::Err => stats.parse_errors += 1,
        }
    }

    stats
}

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: {} <logfile1> [logfile2 ...]", args[0]);
        process::exit(1);
    }

    let file_paths: Vec<&String> = args[1..].iter().collect();

    let mut total_files: u64 = 0;
    let mut total_lines: u64 = 0;
    let mut total_parse_errors: u64 = 0;

    for path_str in &file_paths {
        let content = match fs::read_to_string(path_str) {
            Ok(c) => c,
            Err(e) => {
                eprintln!("Error reading file '{}': {}", path_str, e);
                process::exit(1);
            }
        };

        let all_lines: Vec<&str> = content.lines().collect();
        let stats = aggregate_stats(&all_lines);

        let filename = Path::new(path_str)
            .file_name()
            .and_then(|f| f.to_str())
            .unwrap_or(path_str.as_str());

        println!(
            "{}: ERROR={} WARN={} INFO={} DEBUG={} parse_errors={}",
            filename, stats.error, stats.warn, stats.info, stats.debug, stats.parse_errors
        );

        total_files += 1;
        total_lines += stats.total_lines();
        total_parse_errors += stats.parse_errors;
    }

    println!(
        "TOTAL: files={} lines={} parse_errors={}",
        total_files, total_lines, total_parse_errors
    );

    process::exit(0);
}
