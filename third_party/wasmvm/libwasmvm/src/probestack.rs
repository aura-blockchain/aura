#![allow(dead_code)]

// Provide a local implementation of `__rust_probestack` so that dynamic
// libraries built with modern toolchains still export the symbol the wasm
// runtime expects. This mirrors the upstream compiler-builtins logic for
// Linux x86-64 and ensures stack probes remain effective.
#[cfg(all(target_arch = "x86_64", target_os = "linux"))]
core::arch::global_asm!(
    r#"
    .globl __rust_probestack
__rust_probestack:
    pushq %rbp
    movq %rsp, %rbp
    mov    %rax, %r11
    cmp    $0x1000, %r11
    jna    3f
2:
    sub    $0x1000, %rsp
    test   %rsp, 8(%rsp)
    sub    $0x1000, %r11
    cmp    $0x1000, %r11
    ja     2b
3:
    sub    %r11, %rsp
    test   %rsp, 8(%rsp)
    add    %rax, %rsp
    leave
    ret
"#,
    options(att_syntax)
);
