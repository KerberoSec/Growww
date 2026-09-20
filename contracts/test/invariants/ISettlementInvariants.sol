// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

interface ISettlementInvariants {
    struct TokenReserveProof {
        address tokenAddress;
        uint256 totalCirculatingTokens;
        uint256 custodialPhysicalShareCount;
        bytes32 merkleRoot;
    }

    function checkDvPSolvencyInvariant(
        bytes32 tradeId,
        address buyer,
        address seller,
        uint256 tokenUnits,
        uint256 inrPaise
    ) external view returns (bool isConserved);

    function checkSGFWaterfallIntegrity(
        address defaultingMember,
        uint256 totalDeficit,
        uint256 coreSGFFloor
    ) external view returns (bool isHierarchyRespected);

    function checkCorporateActionRebaseIntegrity(
        address tokenAddress,
        uint256 multiplier,
        uint256 divisor,
        uint256 preRebaseSupply
    ) external view returns (bool isSupplyAccurate);
}
